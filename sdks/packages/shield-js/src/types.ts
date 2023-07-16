import React from "react";

export interface Strategy {
  name: string;
  params: any;
  endpoint: string;
}

export interface User {
  id: string;
  name: string;
  email: string;
  createdAt: Date;
  updatedAt: Date;
}

export interface Group {
  id: string;
  name: string;
  slug: string;
  backend: string;
  resoure_type: string;
  createdAt: Date;
  updatedAt: Date;
}

export interface Organization {
  id: string;
  name: string;
  slug: string;
  metadata: Record<string, string>;
  createdAt: Date;
  updatedAt: Date;
}

export interface Project {
  id: string;
  name: string;
  slug: string;
  metadata: Record<string, string>;
  createdAt: Date;
  updatedAt: Date;
}

export interface Role {
  id: string;
  name: string;
  types: string[];
}

export interface ShieldClientOptions {
  endpoint: string;
}

export interface InitialState {
  sessionId?: string | null;
}

export interface ShieldProviderProps {
  endpoint: string;
  children: React.ReactNode;
  initialState?: InitialState;
}
