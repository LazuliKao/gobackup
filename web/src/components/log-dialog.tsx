'use client';

import {
  Dialog,
  DialogTrigger,
  DialogSurface,
  DialogTitle,
  DialogBody,
  DialogContent,
  Button,
  Tooltip,
} from '@fluentui/react-components';
import { lazy, Suspense } from 'react';
import Icon from '@/components/icon';

const LogView = lazy(() => import('@/components/log-view'));

export default function LogDialog() {
  return (
    <Dialog>
      <Tooltip content="View Backup Logs" relationship="label">
        <DialogTrigger disableButtonEnhancement>
          <Button 
            size="medium" 
            appearance="transparent" 
            icon={<Icon name="history" />}
          />
        </DialogTrigger>
      </Tooltip>
      <DialogSurface className="max-w-4xl w-full">
        <DialogBody>
          <DialogTitle>
            <div className="flex items-center gap-2">
              <Icon name="history" className="text-orange-500" />
              <span>Backup Execution Logs</span>
            </div>
          </DialogTitle>
<DialogContent className="overflow-hidden min-h-[500px] flex flex-col p-0">
        <div className="flex-1 bg-gray-900 rounded-lg overflow-hidden mt-4 flex flex-col h-full w-full max-w-full">
          <Suspense fallback={<div className="text-white text-center py-8">Loading...</div>}>
            <LogView />
          </Suspense>
        </div>
      </DialogContent>
        </DialogBody>
      </DialogSurface>
    </Dialog>
  );
}
