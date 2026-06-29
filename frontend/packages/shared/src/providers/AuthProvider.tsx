import React, { createContext, useContext, useState, useCallback, useEffect } from 'react';
import type { ReactNode } from 'react';
import type { EnrichedReport } from '../types';
import type { IAuthRepository, LoginResult } from '../domain/repositories';

const TOKEN_KEY = 'lapor-bot-token';
const PROFILE_KEY = 'lapor-bot-profile';

function loadToken(): string | null {
	try {
		return localStorage.getItem(TOKEN_KEY);
	} catch {
		return null;
	}
}

function loadProfile(): EnrichedReport | null {
	try {
		const raw = localStorage.getItem(PROFILE_KEY);
		return raw ? JSON.parse(raw) : null;
	} catch {
		return null;
	}
}

function saveSession(token: string, profile: EnrichedReport): void {
	try {
		localStorage.setItem(TOKEN_KEY, token);
		localStorage.setItem(PROFILE_KEY, JSON.stringify(profile));
	} catch {
		// storage full or unavailable — ignore
	}
}

function clearSession(): void {
	try {
		localStorage.removeItem(TOKEN_KEY);
		localStorage.removeItem(PROFILE_KEY);
	} catch {
		// ignore
	}
}

export interface AuthState {
	token: string | null;
	user: EnrichedReport | null;
	isLoading: boolean;
	login: (phone: string) => Promise<EnrichedReport>;
	logout: () => void;
	getToken: () => string | null;
}

const AuthContext = createContext<AuthState | null>(null);

export interface AuthProviderProps {
	authRepo: IAuthRepository;
	children: ReactNode;
	onUnauthorized?: () => void;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ authRepo, children, onUnauthorized }) => {
	const [token, setToken] = useState<string | null>(loadToken);
	const [user, setUser] = useState<EnrichedReport | null>(loadProfile);
	const [isLoading, setIsLoading] = useState(false);

	const getToken = useCallback(() => token, [token]);

	const login = useCallback(async (phone: string): Promise<EnrichedReport> => {
		setIsLoading(true);
		try {
			const result: LoginResult = await authRepo.login(phone);
			setToken(result.token);
			setUser(result.profile);
			saveSession(result.token, result.profile);
			return result.profile;
		} finally {
			setIsLoading(false);
		}
	}, [authRepo]);

	const logout = useCallback(() => {
		setToken(null);
		setUser(null);
		clearSession();
		onUnauthorized?.();
	}, [onUnauthorized]);

	// Sync token across tabs
	useEffect(() => {
		const handler = (e: StorageEvent) => {
			if (e.key === TOKEN_KEY) {
				setToken(e.newValue);
				if (!e.newValue) {
					setUser(null);
				}
			}
		};
		window.addEventListener('storage', handler);
		return () => window.removeEventListener('storage', handler);
	}, []);

	return (
		<AuthContext.Provider value={{ token, user, isLoading, login, logout, getToken }}>
			{children}
		</AuthContext.Provider>
	);
};

export const useAuthContext = (): AuthState => {
	const ctx = useContext(AuthContext);
	if (!ctx) {
		throw new Error('useAuthContext must be used within an AuthProvider');
	}
	return ctx;
};
