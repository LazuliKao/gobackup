'use client';

import { Button, Skeleton, Spinner } from '@fluentui/react-components';
import { filesize } from 'filesize';
import { useEffect, useState } from 'react';
import Link from 'next/link';
import { api, FileItem } from '../../../lib/api';

const Icon = ({
  name,
  mode = 'line',
  className = '',
  loading = false,
}: {
  name: string;
  mode?: 'line' | 'fill';
  className?: string;
  loading?: boolean;
}) => {
  let classes = `ricon ri-${name}-${mode} ${className}`;
  if (loading) {
    classes += ' ricon-loading';
  }
  return <i className={classes}></i>;
};

const PageTitle = ({
  title,
  backTo = '/',
  extra,
}: {
  title: string | React.ReactNode;
  backTo?: string;
  extra?: React.ReactNode;
}) => {
  return (
    <div className="flex items-center space-x-3 pb-3 justify-between">
      <div className="flex items-center gap-3">
        <Link
          href={backTo}
          className="text-2xl hover:text-red rounded hover:border-gray-100"
        >
          <Icon name="arrow-left" />
        </Link>
        <div className="text-xl">{title}</div>
      </div>
      {extra && <div>{extra}</div>}
    </div>
  );
};

const Time = ({ value }: { value: string }) => {
  if (!value) return null;
  return <span title={value}>{new Date(value).toLocaleString()}</span>;
};

const FileItemRow = ({
  file,
  modelName,
}: {
  file: FileItem;
  modelName: string;
}) => {
  const downloadURL = api.getDownloadUrl(modelName, file.filename);
  const fsize = filesize(file.size || 0, { base: 2 }).toString();

  return (
    <div className="flex flex-col gap-2 py-2 px-2 hover:bg-gray-50">
      <a
        className="flex items-center space-x-2 hover:text-blue"
        href={downloadURL}
      >
        <Icon name="folder-zip" />
        <div className="truncate">{file.filename}</div>
      </a>
      <div className="flex flex-col md:flex-row md:items-center md:justify-between text-sm space-y-1 md:space-y-0 md:space-x-4 text-gray-400">
        <div className="flex items-center space-x-4">
          <div>{fsize}</div>
          <div>
            <Time value={file.last_modified || ''} />
          </div>
        </div>
        <div>
          <Button size="small" title="Download backup file.">
            <a href={downloadURL}>
              <Icon name="download-cloud" mode="fill" />
            </a>
          </Button>
        </div>
      </div>
    </div>
  );
};

export default function BrowserClient({ model }: { model?: string[] }) {
  const modelName = model?.[0] || '';
  const [loading, setLoading] = useState(true);
  const [files, setFiles] = useState<FileItem[]>([]);

  const reloadList = async () => {
    if (!modelName) return;
    setLoading(true);
    try {
      const data = await api.listFiles(modelName, '/');
      setFiles(data.files || []);
    } catch (error) {
      console.error('Failed to load files:', error);
      setFiles([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (modelName) {
      reloadList();
    }
  }, [modelName]);

  if (!modelName) {
    return (
      <div className="p-4 text-center text-gray-400">
        No model specified. <Link href="/" className="text-blue">Go back to dashboard</Link>
      </div>
    );
  }

  return (
    <div className="p-4">
      <PageTitle
        title={
          <div className="flex lg:items-center flex-col lg:flex-row-reverse lg:gap-x-2">
            <div className="text-xs text-gray-600">Browser</div>
            <div className="uppercase text-base">{modelName}</div>
          </div>
        }
        backTo={`/`}
        extra={
          <Button size="small" onClick={reloadList} title="Refresh">
            {loading ? <Spinner size="tiny" /> : <Icon name="refresh" />}
          </Button>
        }
      />
      <div className="rounded overflow-y-scroll border border-gray-200 shadow-sm divide-y divide-gray-100 p-2">
        {loading && <Skeleton />}
        {!loading && (
          <>
            {files.length === 0 && (
              <div className="text-center py-10 text-gray-400">
                No backup files found
              </div>
            )}
            {files.map((file, i) => (
              <FileItemRow key={i} file={file} modelName={modelName} />
            ))}
          </>
        )}
      </div>
    </div>
  );
}
