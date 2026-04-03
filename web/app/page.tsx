'use client';

import {
  Button,
  Skeleton,
  Dialog,
  DialogTrigger,
  DialogSurface,
  DialogTitle,
  DialogBody,
  DialogContent,
  DialogActions,
  useToastController,
  Toast,
  ToastTitle,
  ToastBody,
  Toaster,
} from '@fluentui/react-components';
import { useEffect, useState } from 'react';
import dynamic from 'next/dynamic';
import Link from 'next/link';
import { api, ModelConfig } from '@/lib/api';
import Icon from '@/components/icon';

const LogView = dynamic(() => import('@/components/log-view'), { ssr: false });

const ModelList = () => {
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

  const ModelItem = ({ modelKey }: { modelKey: string }) => {
    const model = models[modelKey];
    const scheduleEnable = model.schedule?.enabled;

    return (
      <div className="model-list-item">
        <div className="text-base">
          <div className="text-base font-medium uppercase">{modelKey}</div>
          {scheduleEnable && (
            <div className="text-green text-sm">{model.schedule_info}</div>
          )}
          {model.description && (
            <div className="text-gray-400 truncate text-xs my-1">
              {model.description}
            </div>
          )}
        </div>
        <div className="flex items-center space-x-1">
          <Link href={`/browser/${modelKey}`}>
            <Button size="small">
              <Icon name="folders" />
            </Button>
          </Link>

          <Dialog>
            <DialogTrigger disableButtonEnhancement>
              <Button size="small" title="Perform backup now!">
                <Icon name="play" mode="fill" />
              </Button>
            </DialogTrigger>
            <DialogSurface>
              <DialogBody>
                <DialogTitle>Perform Backup</DialogTitle>
                <DialogContent>
                  Are you sure to perform backup now?
                </DialogContent>
                <DialogActions>
                  <DialogTrigger disableButtonEnhancement>
                    <Button appearance="secondary">Cancel</Button>
                  </DialogTrigger>
                  <DialogTrigger disableButtonEnhancement>
                    <Button
                      appearance="primary"
                      onClick={() => performBackup(modelKey)}
                    >
                      Confirm
                    </Button>
                  </DialogTrigger>
                </DialogActions>
              </DialogBody>
            </DialogSurface>
          </Dialog>
        </div>
      </div>
    );
  };

  return (
    <div className="model-list-wrapper">
      <div className="model-list-header">
        <div className="flex items-center space-x-2">
          <Icon name="stack" />
          <div className="text-text text-base">Models</div>
        </div>
      </div>
      <div className="model-list-scrollview">
        {loading && (
          <div className="p-4">
            <Skeleton />
          </div>
        )}
        {!loading && (
          <>
            {Object.keys(models).map((key: string, idx: number) => (
              <ModelItem modelKey={key} key={idx} />
            ))}
          </>
        )}
      </div>
    </div>
  );
};

export default function Home() {
  return (
    <div className="flex flex-col relative md:flex-row gap-4">
      <ModelList />
      <LogView />
    </div>
  );
}
