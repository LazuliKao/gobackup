'use client';

import {
  Dialog,
  DialogTrigger,
  DialogSurface,
  DialogTitle,
  DialogBody,
  DialogContent,
  Button,
} from '@fluentui/react-components';
import dynamic from 'next/dynamic';
import Icon from '@/components/icon';

const LogView = dynamic(() => import('@/components/log-view'), { ssr: false });

export default function LogDialog() {
  return (
    <Dialog>
      <DialogTrigger disableButtonEnhancement>
        <Button size="small" title="View Backup Logs">
          <Icon name="scroll" />
        </Button>
      </DialogTrigger>
      <DialogSurface>
        <DialogBody>
          <DialogTitle>Backup Logs</DialogTitle>
          <DialogContent>
            <div className="min-h-[400px]">
              <LogView />
            </div>
          </DialogContent>
        </DialogBody>
      </DialogSurface>
    </Dialog>
  );
}
