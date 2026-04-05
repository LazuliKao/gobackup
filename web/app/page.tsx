'use client';

import {
  Button,
  useToastController,
  Toast,
  ToastTitle,
  ToastBody,
  Toaster,
  TabList,
  Tab,
} from '@fluentui/react-components';
import { useEffect, useState } from 'react';
import Link from 'next/link';
import { api, ModelConfig } from '@/lib/api';
import Icon from '@/components/icon';
import ModelGrid from '@/components/model-grid';
import LogDialog from '@/components/log-dialog';

export default function Home() {
  const [selectedTab, setSelectedTab] = useState('dashboard');
  const [loading, setLoading] = useState(false);
  const [models, setModels] = useState<Record<string, ModelConfig>>({});
  const { dispatchToast } = useToastController();

  useEffect(() => {
    reloadModels();
  }, []);

  const performBackup = (model: string) => {
    api.performBackup(model)
      .then(() => {
        dispatchToast(
          <Toast>
            <ToastTitle>Backup</ToastTitle>
            <ToastBody>Backup for {model} performed successfully.</ToastBody>
          </Toast>,
          { intent: 'success' }
        );
      })
      .catch((data) => {
        dispatchToast(
          <Toast>
            <ToastTitle>Backup Failed</ToastTitle>
            <ToastBody>{data.message}</ToastBody>
          </Toast>,
          { intent: 'error' }
        );
      });
  };

  const reloadModels = () => {
    setLoading(true);
    api.getConfig()
      .then((data) => {
        setModels(data.models);
        setLoading(false);
      });
  };

  return (
    <div className="flex flex-col gap-4">
      <TabList
        selectedValue={selectedTab}
        onTabSelect={(_, data) => setSelectedTab(data.value as string)}
      >
        <Tab value="dashboard">Dashboard</Tab>
        <Tab value="browser">Browser</Tab>
      </TabList>

      {selectedTab === 'dashboard' && (
        <div className="flex flex-col gap-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <Icon name="stack" />
              <div className="text-text text-base font-medium">Models</div>
            </div>
            <LogDialog />
          </div>
          <ModelGrid
            models={models}
            loading={loading}
            onBackup={performBackup}
          />
        </div>
      )}

      {selectedTab === 'browser' && (
        <div className="flex items-center justify-center min-h-[300px]">
          <div className="text-center">
            <div className="text-gray-400 text-base mb-4">
              Select a model to browse backup files
            </div>
            <Link
              href="/browser"
              className="text-blue-500 hover:underline"
            >
              Go to Browser
            </Link>
          </div>
        </div>
      )}
    </div>
  );
}
