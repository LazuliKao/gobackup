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
import dynamic from 'next/dynamic';
import Icon from '@/components/icon';

const LogView = dynamic(() => import('@/components/log-view'), { ssr: false });

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
          <DialogContent>
            <div className="min-h-[500px] bg-gray-900 rounded-lg overflow-hidden mt-4">
              <LogView />
            </div>
          </DialogContent>
        </DialogBody>
      </DialogSurface>
    </Dialog>
  );
}
