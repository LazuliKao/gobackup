const API_URL = '/api';

export interface ConfigFileMessageResponse {
  message: string;
}

export class ApiError extends Error {
  status: number;

  payload?: ConfigFileMessageResponse;

  constructor(status: number, message: string, payload?: ConfigFileMessageResponse) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.payload = payload;
  }
}

const fallbackApiErrorMessage = (status: number): string => {
  return `Request failed with status ${status}`;
};

const readApiMessage = async (res: Response): Promise<ConfigFileMessageResponse> => {
  const text = await res.text();

  if (!text) {
    return { message: fallbackApiErrorMessage(res.status) };
  }

  try {
    const data = JSON.parse(text) as Partial<ConfigFileMessageResponse>;
    if (typeof data.message === 'string' && data.message.trim()) {
      return { message: data.message };
    }
  } catch {
    // Fall through to raw text/fallback handling.
  }

  const trimmed = text.trim();
  return {
    message: trimmed || fallbackApiErrorMessage(res.status),
  };
};

const throwApiError = async (res: Response): Promise<never> => {
  const payload = await readApiMessage(res);
  throw new ApiError(res.status, payload.message, payload);
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

export interface ConfigFileResponse {
  yaml: string;
}

export const api = {
  getConfig: async (): Promise<ConfigResponse> => {
    const res = await fetch(`${API_URL}/config`);
    return res.json();
  },

  getConfigFile: async (): Promise<string> => {
    const res = await fetch(`${API_URL}/config/file`);

    if (!res.ok) {
      await throwApiError(res);
    }

    return res.text();
  },

  loadConfigFile: async (): Promise<string> => {
    return api.getConfigFile();
  },

  saveConfigFile: async (yaml: string): Promise<ConfigFileMessageResponse> => {
    const res = await fetch(`${API_URL}/config/file`, {
      method: 'POST',
      headers: {
        'Content-Type': 'text/yaml',
      },
      body: yaml,
    });

    if (!res.ok) {
      await throwApiError(res);
    }

    return res.json();
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
