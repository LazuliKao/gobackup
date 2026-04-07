'use client';

import {
  Button,
  useToastController,
  Toast,
  ToastTitle,
  ToastBody,
  TabList,
  Tab,
} from '@fluentui/react-components';
import { useEffect, useState, ReactNode } from 'react';
import { api, ModelConfig } from '@/lib/api';
import Icon from '@/components/icon';
import ModelGrid from '@/components/model-grid';
import LogDialog from '@/components/log-dialog';
import BrowserClient from '@/app/browser/browser-client';

export default function Home() {
  const [selectedTab, setSelectedTab] = useState('dashboard');
  const [browsingModel, setBrowsingModel] = useState<string | null>(null);
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
            <ToastTitle>Backup Started</ToastTitle>
            <ToastBody>Backup for <strong>{model}</strong> is performing in the background.</ToastBody>
          </Toast> as ReactNode,
          { intent: 'success' }
        );
      })
      .catch((data) => {
        dispatchToast(
          <Toast>
            <ToastTitle>Backup Failed</ToastTitle>
            <ToastBody>{data.message || 'Unknown error occurred'}</ToastBody>
          </Toast> as ReactNode,
          { intent: 'error' }
        );
      });
  };

  const handleBrowse = (model: string) => {
    setBrowsingModel(model);
    setSelectedTab('browser');
  };

  const reloadModels = () => {
    setLoading(true);
    api.getConfig()
      .then((data) => {
        setModels(data.models || {});
      })
      .catch((err) => {
        console.error('Failed to load config:', err);
      })
      .finally(() => {
        setLoading(false);
      });
  };

  return (
    <div className="max-w-6xl mx-auto flex flex-col gap-8 py-6 px-4">
      <div className="flex items-center justify-between border-b pb-4 border-gray-100">
        <TabList
          selectedValue={selectedTab}
          onTabSelect={(_, data) => setSelectedTab(data.value as string)}
          appearance="subtle"
        >
          <Tab value="dashboard" icon={<Icon name="dashboard" />}>Dashboard</Tab>
          <Tab value="browser" icon={<Icon name="folder-zip" />}>Browser</Tab>
        </TabList>
        
        {selectedTab === 'dashboard' && (
          <div className="flex items-center gap-2">
             <Button 
               size="medium" 
               appearance="transparent" 
               onClick={reloadModels}
               icon={<Icon name="refresh" className={loading ? 'animate-spin' : ''} />}
             />
             <LogDialog />
          </div>
        )}
      </div>

      <div className="animate-in fade-in duration-500">
          {selectedTab === 'dashboard' && (
            <div className="flex flex-col gap-6">
              <div className="flex items-center gap-3">
                 <div className="p-2 bg-orange-100 rounded-lg">
                    <Icon name="stack" className="text-orange-600 text-xl" />
                 </div>
                 <div>
                    <h1 className="text-xl font-bold text-gray-800">Backup Models</h1>
                    <p className="text-sm text-gray-500 font-normal">Manage and monitor your backup tasks</p>
                 </div>
              </div>
              <ModelGrid
                models={models}
                loading={loading}
                onBackup={performBackup}
                onBrowse={handleBrowse}
              />
            </div>
          )}

          {selectedTab === 'browser' && (
            <div>
              {browsingModel ? (
                <BrowserClient 
                  model={[browsingModel]} 
                  onBack={() => setSelectedTab('dashboard')} 
                />
              ) : (
                <div className="flex flex-col items-center justify-center min-h-[400px] bg-slate-50 rounded-2xl border-2 border-dashed border-slate-200">
                  <div className="text-center max-w-sm">
                    <div className="bg-white p-4 rounded-full shadow-sm inline-block mb-6">
                       <Icon name="folder-search" className="text-5xl text-blue-500" />
                    </div>
                    <h2 className="text-xl font-semibold text-gray-800 mb-2">Explore Backups</h2>
                    <p className="text-gray-500 text-sm mb-6 leading-relaxed">
                      Select a model from the dashboard to browse and download your backup files.
                    </p>
                    <Button 
                        appearance="primary" 
                        size="large"
                        onClick={() => setSelectedTab('dashboard')}
                    >
                        Select a Model
                    </Button>
                  </div>
                </div>
              )}
            </div>
          )}
      </div>
    </div>
  );
}
