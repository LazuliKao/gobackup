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
import { useCallback, useEffect, useState, ReactNode } from 'react';
import { api, ModelConfig } from '@/lib/api';
import ConfigEditor from '@/components/config-editor';
import Icon from '@/components/icon';
import ModelGrid from '@/components/model-grid';
import LogDialog from '@/components/log-dialog';
import BrowserClient from '@/app/browser/browser-client';

export default function Home() {
  const [selectedTab, setSelectedTab] = useState('dashboard');
  const [browsingModel, setBrowsingModel] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [models, setModels] = useState<Record<string, ModelConfig>>({});
  const [configValue, setConfigValue] = useState('');
  const [configLoading, setConfigLoading] = useState(false);
  const [configSaving, setConfigSaving] = useState(false);
  const [configValid, setConfigValid] = useState(true);
  const [configError, setConfigError] = useState<string | null>(null);
  const { dispatchToast } = useToastController();

  useEffect(() => {
    reloadModels();
  }, []);

  const loadConfigValue = useCallback(() => {
    setConfigLoading(true);
    setConfigError(null);

    api.loadConfigFile()
      .then((yaml) => {
        setConfigValue(yaml);
      })
      .catch((error) => {
        const message = error instanceof Error
          ? error.message
          : 'Failed to load the active configuration file.';

        setConfigError(message);
        setConfigValue('');
      })
      .finally(() => {
        setConfigLoading(false);
      });
  }, []);

  const saveConfigValue = useCallback(() => {
    if (!configValid) {
      dispatchToast(
        <Toast>
          <ToastTitle>Configuration Invalid</ToastTitle>
          <ToastBody>Fix YAML validation issues before saving.</ToastBody>
        </Toast> as ReactNode,
        { intent: 'error' }
      );
      return;
    }

    if (configLoading || configSaving) {
      return;
    }

    setConfigSaving(true);

    api.saveConfigFile(configValue)
      .then((response) => {
        dispatchToast(
          <Toast>
            <ToastTitle>Configuration Saved</ToastTitle>
            <ToastBody>{response.message || 'Configuration file saved.'}</ToastBody>
          </Toast> as ReactNode,
          { intent: 'success' }
        );
      })
      .catch((error) => {
        dispatchToast(
          <Toast>
            <ToastTitle>Configuration Save Failed</ToastTitle>
            <ToastBody>{error instanceof Error ? error.message : 'Failed to save configuration.'}</ToastBody>
          </Toast> as ReactNode,
          { intent: 'error' }
        );
      })
      .finally(() => {
        setConfigSaving(false);
      });
  }, [configLoading, configSaving, configValid, configValue, dispatchToast]);

  useEffect(() => {
    if (selectedTab !== 'config' || configLoading || Boolean(configValue) || Boolean(configError)) {
      return;
    }

    loadConfigValue();
  }, [configError, configLoading, configValue, loadConfigValue, selectedTab]);

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
          <Tab value="config" icon={<Icon name="file-code" />}>Config</Tab>
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
        {selectedTab === 'config' && (
          <Button
            appearance="transparent"
            size="medium"
            onClick={loadConfigValue}
            disabled={configLoading || configSaving}
            icon={<Icon name="refresh" className={configLoading || configSaving ? 'animate-spin' : ''} />}
          >
            Reload
          </Button>
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

          {selectedTab === 'config' && (
            <div className="flex flex-col gap-6">
              <div className="flex items-center gap-3">
                <div className="rounded-lg bg-orange-100 p-2">
                  <Icon name="file-code" className="text-xl text-orange-600" />
                </div>
                <div>
                  <h1 className="text-xl font-bold text-gray-800">Edit Configuration</h1>
                  <p className="text-sm font-normal text-gray-500">
                    Edit the active gobackup.yml with inline YAML validation and explicit save controls.
                  </p>
                </div>
              </div>

              {configError ? (
                <div className="rounded-2xl border border-red-200 bg-red-50 px-6 py-5 shadow-sm">
                  <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
                    <div className="flex items-start gap-3">
                      <div className="rounded-xl bg-white p-2 text-red-500 shadow-sm">
                        <Icon name="alert" className="text-xl" />
                      </div>
                      <div>
                        <h2 className="text-base font-semibold text-red-800">Unable to load configuration</h2>
                        <p className="mt-1 text-sm text-red-700">{configError}</p>
                      </div>
                    </div>

                    <Button appearance="secondary" onClick={loadConfigValue}>
                      Try Again
                    </Button>
                  </div>
                </div>
              ) : (
                <ConfigEditor
                  value={configValue}
                  onChange={setConfigValue}
                  loading={configLoading || configSaving}
                  disabled={configSaving}
                  onSave={saveConfigValue}
                  onValidationChange={(isValid) => setConfigValid(isValid)}
                  height={560}
                />
              )}
            </div>
          )}
      </div>
    </div>
  );
}
