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
  Tooltip,
} from '@fluentui/react-components';
import { ModelConfig } from '@/lib/api';
import Icon from '@/components/icon';

export interface ModelCardProps {
  modelKey: string;
  model: ModelConfig;
  onBackup: (key: string) => void;
  onBrowse: (key: string) => void;
}

export default function ModelCard({
  modelKey,
  model,
  onBackup,
  onBrowse,
}: ModelCardProps) {
  const scheduleEnable = model.schedule?.enabled;

  return (
    <Card appearance="filled-alternative" className="hover:shadow-md transition-shadow">
      <CardHeader
        header={
          <div className="flex-1">
            <div className="text-base font-semibold uppercase text-gray-800">{modelKey}</div>
            {scheduleEnable && (
              <div className="text-green-600 text-xs font-medium mt-1">
                <Icon name="time" className="mr-1 inline-block align-text-bottom" />
                {model.schedule_info}
              </div>
            )}
          </div>
        }
      />
      
      <div className="px-4 py-3 flex-grow">
        {model.description ? (
          <div className="text-gray-500 text-sm italic">{model.description}</div>
        ) : (
          <div className="text-gray-300 text-sm">No description available</div>
        )}
      </div>

      <CardFooter>
        <div className="flex items-center space-x-2">
          <Tooltip content="Browse backup files" relationship="label">
            <Button 
              size="medium" 
              appearance="subtle" 
              icon={<Icon name="folder-open" />} 
              onClick={() => onBrowse(modelKey)}
            />
          </Tooltip>

          <Dialog>
            <Tooltip content="Perform backup now" relationship="label">
              <DialogTrigger disableButtonEnhancement>
                <Button size="medium" appearance="subtle" icon={<Icon name="play" mode="fill" />} />
              </DialogTrigger>
            </Tooltip>
            <DialogSurface>
              <DialogBody>
                <DialogTitle>Perform Backup</DialogTitle>
                <DialogContent>
                  Are you sure you want to perform a backup for <strong>{modelKey}</strong> now?
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
