export function getJobColor(jobId: string): string {
  switch (jobId?.toLowerCase()) {
    case "fighter": return "text-system-red bg-system-red/10 border-system-red/30";
    case "tank": return "text-system-gold bg-system-gold/10 border-system-gold/30";
    case "assassin": return "text-system-purple bg-system-purple/10 border-system-purple/30";
    case "mage": return "text-red-400 bg-red-400/10 border-red-400/30";
    case "ranger": return "text-system-blue bg-system-blue/10 border-system-blue/30";
    case "healer": return "text-system-green bg-system-green/10 border-system-green/30";
    case "necromancer": return "text-gray-400 bg-gray-800/40 border-gray-600/30";
    default: return "text-gray-400 bg-gray-800/30 border-gray-700/30";
  }
}

export function getJobBadgeClass(jobId: string): string {
  switch (jobId?.toLowerCase()) {
    case "fighter": return "text-system-red bg-system-red/5 border-system-red/20";
    case "tank": return "text-system-gold bg-system-gold/5 border-system-gold/20";
    case "assassin": return "text-system-purple bg-system-purple/5 border-system-purple/20";
    case "mage": return "text-red-400 bg-red-400/5 border-red-400/20";
    case "ranger": return "text-system-blue bg-system-blue/5 border-system-blue/20";
    case "healer": return "text-system-green bg-system-green/5 border-system-green/20";
    case "necromancer": return "text-gray-400 bg-gray-800/20 border-gray-600/20";
    default: return "text-gray-400 bg-gray-800/20 border-gray-700/20";
  }
}
