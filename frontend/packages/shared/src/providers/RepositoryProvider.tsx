import React, { createContext, useContext } from 'react';
import type { ReactNode } from 'react';
import type { IReportRepository, IAuthRepository } from '../domain/repositories';

export interface RepositoryContextType {
  reports: IReportRepository;
  auth: IAuthRepository;
}

const RepositoryContext = createContext<RepositoryContextType | null>(null);

export interface RepositoryProviderProps {
  repositories: RepositoryContextType;
  children: ReactNode;
}

export const RepositoryProvider: React.FC<RepositoryProviderProps> = ({ repositories, children }) => {
  return (
    <RepositoryContext.Provider value={repositories}>
      {children}
    </RepositoryContext.Provider>
  );
};

export const useRepositories = (): RepositoryContextType => {
  const context = useContext(RepositoryContext);
  if (!context) {
    throw new Error('useRepositories must be used within a RepositoryProvider');
  }
  return context;
};
