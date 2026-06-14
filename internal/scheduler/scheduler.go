// Package scheduler provides a lightweight, context-aware job scheduler
// for recurring tasks like daily notifications. It uses time.NewTimer
// (not time.After) to avoid timer leaks, supports missed-run recovery,
// and drains cleanly on context cancellation.
package scheduler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// Schedule determines when a job should next run.
type Schedule interface {
	// Next returns the next activation time after (or at) now.
	Next(now time.Time) time.Time
}

// DailySchedule fires once per day at the specified hour and minute
// in the given location. If now is already past that time today,
// Next returns tomorrow's occurrence.
type DailySchedule struct {
	Hour   int
	Minute int
	Loc    *time.Location
}

func (d DailySchedule) Next(now time.Time) time.Time {
	today := time.Date(now.Year(), now.Month(), now.Day(), d.Hour, d.Minute, 0, 0, d.Loc)
	if now.After(today) || now.Equal(today) {
		return today.Add(24 * time.Hour)
	}
	return today
}

// WeeklySchedule fires once per week on the specified weekday, hour, and minute
// in the given location.
type WeeklySchedule struct {
	Weekday time.Weekday
	Hour    int
	Minute  int
	Loc     *time.Location
}

func (w WeeklySchedule) Next(now time.Time) time.Time {
	nowInLoc := now.In(w.Loc)
	target := time.Date(nowInLoc.Year(), nowInLoc.Month(), nowInLoc.Day(), w.Hour, w.Minute, 0, 0, w.Loc)
	
	daysDiff := int(w.Weekday) - int(nowInLoc.Weekday())
	if daysDiff < 0 {
		daysDiff += 7
	}
	target = target.AddDate(0, 0, daysDiff)
	
	if !target.After(nowInLoc) {
		target = target.AddDate(0, 0, 7)
	}
	return target
}

// Job is a single scheduled task.
type Job struct {
	Name    string
	Freq    Schedule
	Fn      func(ctx context.Context) error
	Recover bool // if true, fire immediately when past Next() by >1 minute

	lastRun time.Time
	nextRun time.Time
}

// Scheduler manages a set of recurring jobs.
type Scheduler struct {
	jobs   []*Job
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mu     sync.Mutex
}

// NewScheduler creates a Scheduler that cancels all jobs when ctx is done.
func NewScheduler(ctx context.Context) *Scheduler {
	ctx, cancel := context.WithCancel(ctx)
	return &Scheduler{ctx: ctx, cancel: cancel}
}

// AddJob registers a job. Must be called before Start.
func (s *Scheduler) AddJob(job *Job) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs = append(s.jobs, job)
}

// Start launches one goroutine per job. Each goroutine computes its
// own next activation, handles missed-run recovery (if Recover is true),
// and stops when the scheduler context is cancelled.
func (s *Scheduler) Start() {
	s.mu.Lock()
	jobs := make([]*Job, len(s.jobs))
	copy(jobs, s.jobs)
	s.mu.Unlock()

	for _, job := range jobs {
		job := job
		s.wg.Add(1)
		go s.runJob(job)
	}
}

// Stop cancels the scheduler context and waits for all job goroutines
// to finish. Safe to call multiple times.
func (s *Scheduler) Stop() {
	s.cancel()
	s.wg.Wait()
}

func (s *Scheduler) runJob(job *Job) {
	defer s.wg.Done()

	// Initial scheduling
	now := time.Now()
	job.nextRun = job.Freq.Next(now)

	// Recovery: if we missed today's scheduled run by more than 1 minute,
	// fire immediately. job.nextRun is tomorrow's occurrence, so today's
	// was exactly 24h earlier.
	if job.Recover {
		missedRun := job.nextRun.Add(-24 * time.Hour)
		missedBy := now.Sub(missedRun)
		if missedBy > 1*time.Minute {
			log.Printf("[SCHEDULER] %s: missed run by %v, recovering now", job.Name, missedBy.Round(time.Second))
			s.executeJob(job)
			job.nextRun = job.Freq.Next(time.Now())
		}
	}

	for {
		delay := time.Until(job.nextRun)
		log.Printf("[SCHEDULER] %s: next run at %v (in %v)", job.Name, job.nextRun, delay.Round(time.Second))

		timer := time.NewTimer(delay)
		select {
		case <-timer.C:
			s.executeJob(job)
			job.nextRun = job.Freq.Next(time.Now())
		case <-s.ctx.Done():
			timer.Stop()
			log.Printf("[SCHEDULER] %s: cancelled", job.Name)
			return
		}
	}
}

func (s *Scheduler) executeJob(job *Job) {
	log.Printf("[SCHEDULER] %s: executing", job.Name)
	if err := job.Fn(s.ctx); err != nil {
		log.Printf("[SCHEDULER] %s: error: %v", job.Name, err)
		if err == context.Canceled {
			return
		}
	}
	job.lastRun = time.Now()
}

// parseDaily parses an "HH:MM" string into a DailySchedule.
// Returns an error if the format is invalid.
func ParseDaily(timeStr string, loc *time.Location) (*DailySchedule, error) {
	var h, m int
	_, err := fmt.Sscanf(timeStr, "%d:%d", &h, &m)
	if err != nil {
		return nil, fmt.Errorf("invalid time format %q: expected HH:MM", timeStr)
	}
	if h < 0 || h > 23 || m < 0 || m > 59 {
		return nil, fmt.Errorf("invalid time %q: hour 0-23, minute 0-59", timeStr)
	}
	return &DailySchedule{Hour: h, Minute: m, Loc: loc}, nil
}
