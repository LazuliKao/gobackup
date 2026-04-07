const API_URL = '/api';

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

export const api = {
  getConfig: async (): Promise<ConfigResponse> => {
    const res = await fetch(`${API_URL}/config`);
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
