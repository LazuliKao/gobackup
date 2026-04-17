const API_URL = '/api';

interface ApiErrorPayload {
  message?: string;
}

export class ApiError extends Error {
  status: number;

  constructor(message: string, status: number) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }
}

const getErrorMessage = async (res: Response): Promise<string> => {
  const contentType = res.headers.get('content-type') || '';

  if (contentType.includes('application/json')) {
    try {
      const payload = await res.json() as ApiErrorPayload;
      if (payload.message) {
        return payload.message;
      }
    } catch {
      // Ignore parse failures and fall back to a generic message.
    }
  }

  try {
    const text = await res.text();
    if (text.trim()) {
      return text.trim();
    }
  } catch {
    // Ignore parse failures and fall back to a generic message.
  }

  return `Request failed with status ${res.status}`;
};

const requestJSON = async <T>(input: string, init?: RequestInit): Promise<T> => {
  const res = await fetch(input, init);

  if (!res.ok) {
    throw new ApiError(await getErrorMessage(res), res.status);
  }

  return res.json();
};

const requestText = async (input: string, init?: RequestInit): Promise<string> => {
  const res = await fetch(input, init);

  if (!res.ok) {
    throw new ApiError(await getErrorMessage(res), res.status);
  }

  return res.text();
};

const normalizeConfigRequestError = (error: unknown): never => {
  if (error instanceof Error) {
    throw error;
  }

  throw new Error('Request failed before the config API could respond.');
};

const requestConfigJSON = async <T>(input: string, init?: RequestInit): Promise<T> => {
  try {
    return await requestJSON<T>(input, init);
  } catch (error) {
    return normalizeConfigRequestError(error);
  }
};

const requestConfigText = async (input: string, init?: RequestInit): Promise<string> => {
  try {
    return await requestText(input, init);
  } catch (error) {
    return normalizeConfigRequestError(error);
  }
};

export interface ModelConfig {
  description?: string;
  schedule?: {
    enabled: boolean;
  };
  schedule_info?: string;
}

export interface ConfigResponse {
  models: Record<string, ModelConfig>;
}

export interface ConfigPathsResponse {
  paths: string[];
  allowed_paths?: Array<{
    path: string;
    exists: boolean;
  }>;
  current_path?: string;
}

export interface SaveConfigResponse {
  message: string;
}

export interface ValidateConfigResponse {
  message: string;
  valid: boolean;
}

export interface FileItem {
  filename: string;
  size?: number;
  last_modified?: string;
}

export interface ListResponse {
  files: FileItem[];
}

export interface PerformResponse {
  message: string;
}

export const api = {
  getConfig: async (): Promise<ConfigResponse> => {
    const res = await fetch(`${API_URL}/config`);
    return res.json();
  },

  getConfigPaths: async (): Promise<ConfigPathsResponse> => {
    return requestConfigJSON<ConfigPathsResponse>(`${API_URL}/config/paths`);
  },

  getConfigRaw: async (path?: string): Promise<string> => {
    const query = path ? `?${new URLSearchParams({ path }).toString()}` : '';
    return requestConfigText(`${API_URL}/config/raw${query}`);
  },

  saveConfig: async (path: string, content: string, createIfMissing: boolean = false): Promise<SaveConfigResponse> => {
    return requestConfigJSON<SaveConfigResponse>(`${API_URL}/config/save`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ path, content, create_if_missing: createIfMissing }),
    });
  },

  validateConfig: async (path: string, content: string): Promise<ValidateConfigResponse> => {
    return requestConfigJSON<ValidateConfigResponse>(`${API_URL}/config/validate`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ path, content }),
    });
  },

  performBackup: async (model: string): Promise<PerformResponse> => {
    const res = await fetch(`${API_URL}/perform`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ model }),
    });
    return res.json();
  },

  listFiles: async (model: string, parent: string = '/'): Promise<ListResponse> => {
    const query = new URLSearchParams({ model, parent });
    const res = await fetch(`${API_URL}/list?${query.toString()}`);
    return res.json();
  },

  getDownloadUrl: (model: string, path: string): string => {
    return `${API_URL}/download?${new URLSearchParams({ model, path }).toString()}`;
  },

  getLogStreamUrl: (): string => {
    return `${API_URL}/log`;
  },
};
