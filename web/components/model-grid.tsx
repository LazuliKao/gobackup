'use client';

import { Skeleton } from '@fluentui/react-components';
import { ModelConfig } from '@/lib/api';
import ModelCard from '@/components/model-card';

export interface ModelGridProps {
  models: Record<string, ModelConfig>;
  loading: boolean;
  onBackup: (key: string) => void;
}

export default function ModelGrid({
  models,
  loading,
  onBackup,
}: ModelGridProps) {
  const modelKeys = Object.keys(models);

  if (loading) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {Array.from({ length: 6 }).map((_, i) => (
          <div key={i} className="h-48">
            <Skeleton shape="rectangle" />
          </div>
        ))}
      </div>
    );
  }

  if (modelKeys.length === 0) {
    return (
      <div className="flex items-center justify-center min-h-[200px]">
        <div className="text-center text-gray-400">
          <div className="text-base">No models configured</div>
          <div className="text-sm mt-1">
            Add models to your configuration to get started
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      {modelKeys.map((modelKey) => (
        <ModelCard
          key={modelKey}
          modelKey={modelKey}
          model={models[modelKey]}
          onBackup={onBackup}
        />
      ))}
    </div>
  );
}
