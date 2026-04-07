'use client';

import {
  Button,
  Spinner,
  Input,
  Tooltip,
  Table,
  TableHeader,
  TableRow,
  TableHeaderCell,
  TableBody,
  TableCell,
  TableCellLayout,
  Title2,
  Caption1,
  Text,
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbDivider,
  Toolbar,
  ToolbarGroup,
  ToolbarDivider,
} from '@fluentui/react-components';
import { filesize } from 'filesize';
import { FC, useEffect, useState, useMemo } from 'react';
import { api, FileItem } from '../../lib/api';
import Icon from '@/components/icon';

interface BrowserClientProps {
  model?: string[];
  onBack?: () => void;
}

const BrowserClient: FC<BrowserClientProps> = ({ 
  model, 
  onBack 
}) => {
  const modelName = model?.[0] || '';
  const [loading, setLoading] = useState(true);
  const [files, setFiles] = useState<FileItem[]>([]);
  const [searchTerm, setSearchTerm] = useState('');

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

  const filteredFiles = useMemo(() => {
    let result = files.filter((f) =>
      f.filename.toLowerCase().includes(searchTerm.toLowerCase())
    );
    result.sort(
      (a, b) =>
        new Date(b.last_modified || 0).getTime() -
        new Date(a.last_modified || 0).getTime()
    );
    return result;
  }, [files, searchTerm]);

  if (!modelName) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[400px] text-center gap-4">
        <Icon name="error-warning" className="text-4xl text-gray-300" />
        <Title2>No Model Specified</Title2>
        <a href="/">
          <Button appearance="primary">Return to Dashboard</Button>
        </a>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-6">
      {/* Header Section */}
      <div className="flex flex-col gap-4">
        <div className="flex items-center gap-2 text-sm text-gray-500">
          <Button 
              appearance="transparent"
              size="small"
              className="font-normal text-orange-600 hover:text-orange-700 p-0 min-w-0 h-auto"
              onClick={onBack}
          >
             Dashboard
          </Button>
          <Icon name="arrow-right-s" className="text-gray-300 text-xs" />
          <span>Browser</span>
          <Icon name="arrow-right-s" className="text-gray-300 text-xs" />
          <Text weight="semibold" className="uppercase tracking-tight text-gray-800">{modelName}</Text>
        </div>
        
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
             <Button 
                size="medium" 
                appearance="subtle" 
                icon={<Icon name="arrow-left-s" className="text-xl" />} 
                onClick={onBack}
                className="hover:bg-gray-100 rounded-lg p-2"
             />
             <div className="p-2 bg-orange-100 rounded-lg">
                <Icon name="archive-drawer" className="text-orange-600 text-xl" />
             </div>
             <Title2 block className="uppercase tracking-tight">{modelName}</Title2>
          </div>
          <Tooltip content="Refresh backups" relationship="label">
            <Button
              appearance="subtle"
              onClick={reloadList}
              disabled={loading}
              icon={loading ? <Spinner size="tiny" /> : <Icon name="refresh" className="text-gray-500" />}
            />
          </Tooltip>
        </div>
      </div>

      {/* Action Bar */}
      <div className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
        <Toolbar>
          <ToolbarGroup className="flex-1 px-2">
            <Input
              placeholder="Search by filename..."
              contentBefore={<Icon name="search-line" className="text-gray-400" />}
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full max-w-sm"
              appearance="filled-lighter"
            />
          </ToolbarGroup>
          <ToolbarDivider />
          <ToolbarGroup className="px-4">
             <Text size={200} weight="medium" className="text-gray-500 italic">
                {filteredFiles.length} {filteredFiles.length === 1 ? 'archive' : 'archives'} found
             </Text>
          </ToolbarGroup>
        </Toolbar>

        {/* File Table */}
        <div className="overflow-x-auto w-full">
          <Table size="medium">
            <TableHeader className="bg-gray-50">
              <TableRow>
                <TableHeaderCell className="min-w-[120px] sm:min-w-[200px]">Archive Name</TableHeaderCell>
                <TableHeaderCell style={{ width: '100px' }} className="hidden sm:table-cell">Size</TableHeaderCell>
                <TableHeaderCell style={{ width: '150px' }}>Date Modified</TableHeaderCell>
                <TableHeaderCell style={{ width: '60px' }} className="text-center">Action</TableHeaderCell>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={4} className="py-20 text-center">
                    <Spinner label="Scanning backups..." />
                  </TableCell>
                </TableRow>
              ) : filteredFiles.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={4} className="py-20 text-center">
                    <div className="flex flex-col items-center gap-2">
                        <Icon name="inbox-2" className="text-4xl text-gray-200" />
                        <Text className="text-gray-400 italic">No matching backup files were found.</Text>
                    </div>
                  </TableCell>
                </TableRow>
              ) : (
                filteredFiles.map((file) => {
                  const downloadURL = api.getDownloadUrl(modelName, file.filename);
                  return (
                    <TableRow key={file.filename} className="hover:bg-slate-50 transition-colors">
                      <TableCell>
                        <TableCellLayout
                          media={<Icon name="file-zip" className="text-orange-500" />}
                        >
                          <a 
                            href={downloadURL} 
                            target="_blank"
                            className="no-underline hover:text-orange-600 transition-colors font-medium break-all"
                          >
                            {file.filename}
                          </a>
                        </TableCellLayout>
                      </TableCell>
                      <TableCell className="hidden sm:table-cell">
                        <Text size={200} className="font-mono text-gray-500">
                           {filesize(file.size || 0, { base: 2 }).toString()}
                        </Text>
                      </TableCell>
                      <TableCell>
                        <div className="flex flex-col">
                          <Text size={200}>{new Date(file.last_modified || '').toLocaleDateString()}</Text>
                          <Caption1 className="text-gray-400">{new Date(file.last_modified || '').toLocaleTimeString()}</Caption1>
                        </div>
                      </TableCell>
                      <TableCell className="text-center">
                        <Tooltip content="Download archive" relationship="label">
                          <Button
                            appearance="subtle"
                            icon={<Icon name="download-cloud" mode="fill" />}
                            onClick={() => window.open(downloadURL, '_blank')}
                          />
                        </Tooltip>
                      </TableCell>
                    </TableRow>
                  );
                })
              )}
            </TableBody>
          </Table>
        </div>
      </div>
      
      <div className="text-center">
         <Caption1 className="text-gray-400 italic">
            GoBackup provides secure retrieval for all your backup archives.
         </Caption1>
       </div>
     </div>
   );
};

export default BrowserClient;
