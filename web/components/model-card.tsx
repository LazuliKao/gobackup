'use client';

import {
  Card,
  CardHeader,
  CardFooter,
  Button,
  Dialog,
  DialogTrigger,
  DialogSurface,
  DialogTitle,
  DialogBody,
  DialogContent,
  DialogActions,
} from '@fluentui/react-components';
import Link from 'next/link';
import { ModelConfig } from '@/lib/api';
import Icon from '@/components/icon';

export interface ModelCardProps {
  modelKey: string;
  model: ModelConfig;
  onBackup: (key: string) => void;
}

export default function ModelCard({
  modelKey,
  model,
  onBackup,
}: ModelCardProps) {
  const scheduleEnable = model.schedule?.enabled;

  return (
    <Card>
      <CardHeader
        header={
          <div className="flex-1">
            <div className="text-base font-medium uppercase">{modelKey}</div>
            {scheduleEnable && (
              <div className="text-green text-sm">{model.schedule_info}</div>
            )}
          </div>
        }
      />
      {model.description && (
        <div className="px-4 py-2">
          <div className="text-gray-400 text-xs">{model.description}</div>
        </div>
      )}
      <CardFooter>
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
                      onClick={() => onBackup(modelKey)}
                    >
                      Confirm
                    </Button>
                  </DialogTrigger>
                </DialogActions>
              </DialogBody>
            </DialogSurface>
          </Dialog>
        </div>
      </CardFooter>
    </Card>
  );
}
