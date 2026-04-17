'use client';

import Editor from '@monaco-editor/react';
import EditorWorker from 'monaco-editor/esm/vs/editor/editor.worker.js?worker';
import {
  Button,
  Field,
  Select,
  Spinner,
  Toast,
  ToastBody,
  ToastTitle,
  useToastController,
} from '@fluentui/react-components';
import { ChangeEvent, ReactNode, useEffect, useMemo, useState } from 'react';
import Icon from '@/components/icon';
import { api } from '@/lib/api';

interface ConfigPathOption {
  path: string;
  exists: boolean;
}

type MonacoEnvironment = {
  getWorker: (_moduleId: unknown, label: string) => Worker;
};

const monacoGlobal = globalThis as typeof globalThis & {
  MonacoEnvironment?: MonacoEnvironment;
};

if (!monacoGlobal.MonacoEnvironment) {
  monacoGlobal.MonacoEnvironment = {
    getWorker: () => new EditorWorker(),
  };
}

type FeedbackIntent = 'info' | 'success' | 'error';

interface FeedbackMessage {
  intent: FeedbackIntent;
  text: string;
}

const DEFAULT_MESSAGE: FeedbackMessage = {
  intent: 'info',
  text: 'Choose a configuration path to load its raw YAML content in Monaco, then validate or save explicitly when you are ready.',
};

const PATHS_LOADING_MESSAGE: FeedbackMessage = {
  intent: 'info',
  text: 'Loading available configuration paths...',
};

const PATHS_EMPTY_MESSAGE: FeedbackMessage = {
  intent: 'info',
  text: 'No existing configuration files were found in the known GoBackup locations yet.',
};

const CREATE_CONFIRMATION_MESSAGE = 'This config file does not exist yet. Create it now and save your YAML to that allowlisted path?';

export default function ConfigEditor() {
  const [selectedPath, setSelectedPath] = useState<string | null>(null);
  const [availablePaths, setAvailablePaths] = useState<ConfigPathOption[]>([]);
  const [content, setContent] = useState('');
  const [dirty, setDirty] = useState(false);
  const [saving, setSaving] = useState(false);
  const [validating, setValidating] = useState(false);
  const [loadingPaths, setLoadingPaths] = useState(true);
  const [loadingContent, setLoadingContent] = useState(false);
  const [message, setMessage] = useState<FeedbackMessage>(DEFAULT_MESSAGE);
  const { dispatchToast } = useToastController();

  const selectedPathOption = useMemo(
    () => availablePaths.find((option) => option.path === selectedPath) ?? null,
    [availablePaths, selectedPath]
  );

  const existingPaths = useMemo(
    () => availablePaths.filter((option) => option.exists),
    [availablePaths]
  );

  const missingPaths = useMemo(
    () => availablePaths.filter((option) => !option.exists),
    [availablePaths]
  );

  const editorOptions = useMemo(
    () => ({
      automaticLayout: true,
      fontSize: 13,
      lineNumbersMinChars: 3,
      minimap: { enabled: false },
      padding: { top: 16, bottom: 16 },
      readOnly: loadingContent || saving || validating,
      roundedSelection: false,
      scrollBeyondLastLine: false,
      wordWrap: 'on' as const,
    }),
    [loadingContent, saving, validating]
  );

  const getErrorMessage = (error: unknown) => {
    if (error instanceof Error && error.message) {
      return error.message;
    }

    return 'Unknown error occurred.';
  };

  const showToast = (intent: 'success' | 'error', title: string, body: ReactNode) => {
    dispatchToast(
      <Toast>
        <ToastTitle>{title}</ToastTitle>
        <ToastBody>{body}</ToastBody>
      </Toast> as ReactNode,
      { intent }
    );
  };

  useEffect(() => {
    let cancelled = false;

    const loadPaths = async () => {
      setLoadingPaths(true);
      setMessage(PATHS_LOADING_MESSAGE);

      try {
        const data = await api.getConfigPaths();

        if (cancelled) {
          return;
        }

        const existing = data.paths || [];
        const allowed = data.allowed_paths && data.allowed_paths.length > 0
          ? data.allowed_paths
          : existing.map((path) => ({ path, exists: true }));
        const preferredExistingPath = existing.includes(data.current_path || '')
          ? data.current_path || null
          : existing[0] || null;
        const fallbackAllowedPath = allowed[0]?.path || null;
        const initialPath = preferredExistingPath || fallbackAllowedPath;

        setAvailablePaths(allowed);
        setSelectedPath(initialPath);

        if (allowed.length === 0) {
          setContent('');
          setDirty(false);
          setMessage(PATHS_EMPTY_MESSAGE);
          return;
        }

        setMessage({
          intent: 'info',
          text: initialPath
            ? existing.length > 0
              ? `Loaded ${existing.length} existing configuration path${existing.length === 1 ? '' : 's'}. Missing allowlisted paths can be selected and created explicitly when you save.`
              : 'No existing config file was found. Choose an allowlisted path, review your YAML, then confirm creation when you save.'
            : 'Choose a configuration path to load its raw YAML content.',
        });
      } catch (error) {
        console.error('Failed to load config paths:', error);

        if (cancelled) {
          return;
        }

        setAvailablePaths([]);
        setSelectedPath(null);
        setContent('');
        setDirty(false);
        setMessage({
          intent: 'error',
          text: 'Failed to load configuration paths. Make sure the config API is reachable before using the editor.',
        });
      } finally {
        if (!cancelled) {
          setLoadingPaths(false);
        }
      }
    };

    loadPaths();

    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    if (!selectedPath) {
      return;
    }

    if (!selectedPathOption?.exists) {
      setLoadingContent(false);
      setContent('');
      setDirty(false);
      setMessage({
        intent: 'info',
        text: `Selected allowlisted path ${selectedPath}. This file does not exist yet. Add YAML, then click Save Config and confirm creation when prompted.`,
      });
      return;
    }

    let cancelled = false;

    const loadContent = async () => {
      setLoadingContent(true);

      try {
        const rawContent = await api.getConfigRaw(selectedPath);

        if (cancelled) {
          return;
        }

        setContent(rawContent);
        setDirty(false);
        setMessage({
          intent: 'info',
          text: `Loaded configuration from ${selectedPath}. Edit the YAML here, then use Validate YAML or Save Config when you are ready.`,
        });
      } catch (error) {
        console.error('Failed to load config content:', error);

        if (cancelled) {
          return;
        }

        setContent('');
        setDirty(false);
        setMessage({
          intent: 'error',
          text: getErrorMessage(error),
        });
      } finally {
        if (!cancelled) {
          setLoadingContent(false);
        }
      }
    };

    loadContent();

    return () => {
      cancelled = true;
    };
  }, [selectedPath, selectedPathOption?.exists]);

  useEffect(() => {
    const handleBeforeUnload = (event: BeforeUnloadEvent) => {
      if (!dirty) {
        return;
      }

      event.preventDefault();
      event.returnValue = '';
    };

    window.addEventListener('beforeunload', handleBeforeUnload);
    return () => {
      window.removeEventListener('beforeunload', handleBeforeUnload);
    };
  }, [dirty]);

  const handleContentChange = (value?: string) => {
    setContent(value || '');
    setDirty(true);
    setMessage({
      intent: 'info',
      text: selectedPath
        ? `Editing ${selectedPath}. Your changes stay local until you explicitly validate or save them.`
        : DEFAULT_MESSAGE.text,
    });
  };

  const handlePathChange = (event: ChangeEvent<HTMLSelectElement>) => {
    const nextPath = event.currentTarget.value || null;

    if (nextPath === selectedPath) {
      return;
    }

    if (dirty) {
      const shouldDiscardChanges = window.confirm(
        `Discard unsaved changes for ${selectedPath}? Your edits are not saved automatically.`
      );

      if (!shouldDiscardChanges) {
        event.currentTarget.value = selectedPath || '';
        setMessage({
          intent: 'info',
          text: selectedPath
            ? `Still editing ${selectedPath}. Validate or save before switching files.`
            : DEFAULT_MESSAGE.text,
        });
        return;
      }
    }

    setSelectedPath(nextPath);
  };

  const handleValidate = async () => {
    if (!selectedPath || validating || saving || loadingContent) {
      return;
    }

    setValidating(true);

    try {
      const data = await api.validateConfig(selectedPath, content);
      const successMessage = dirty
        ? `${data.message} Unsaved changes are still in the editor until you click Save Config.`
        : `${data.message} There are no unsaved changes right now.`;

      setMessage({
        intent: 'success',
        text: successMessage,
      });
      showToast('success', 'YAML Valid', successMessage);
    } catch (error) {
      const errorMessage = getErrorMessage(error);
      setMessage({
        intent: 'error',
        text: errorMessage,
      });
      showToast('error', 'Validation Failed', errorMessage);
    } finally {
      setValidating(false);
    }
  };

  const handleSave = async () => {
    if (!selectedPath || saving || validating || loadingContent || !dirty) {
      return;
    }

    const isMissingPath = !selectedPathOption?.exists;
    if (isMissingPath) {
      const shouldCreate = window.confirm(`Create a new config file at ${selectedPath}?

${CREATE_CONFIRMATION_MESSAGE}`);
      if (!shouldCreate) {
        setMessage({
          intent: 'info',
          text: `Creation canceled for ${selectedPath}. Existing config files remain preferred until you explicitly confirm creating a new one.`,
        });
        return;
      }
    }

    setSaving(true);

    try {
      const data = await api.saveConfig(selectedPath, content, isMissingPath);
      const successMessage = isMissingPath
        ? `${data.message} Created ${selectedPath}.`
        : `${data.message} Saved ${selectedPath}.`;

      if (isMissingPath) {
        setAvailablePaths((currentPaths) => currentPaths.map((pathOption) => (
          pathOption.path === selectedPath
            ? { ...pathOption, exists: true }
            : pathOption
        )));
      }

      setDirty(false);
      setMessage({
        intent: 'success',
        text: successMessage,
      });
      showToast('success', 'Config Saved', successMessage);
    } catch (error) {
      const errorMessage = getErrorMessage(error);
      setMessage({
        intent: 'error',
        text: errorMessage,
      });
      showToast('error', 'Save Failed', errorMessage);
    } finally {
      setSaving(false);
    }
  };

  const messageClassName = {
    info: 'border-blue-100 bg-blue-50 text-blue-700',
    success: 'border-green-100 bg-green-50 text-green-700',
    error: 'border-red-100 bg-red-50 text-red-700',
  }[message.intent];

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center gap-3">
        <div className="p-2 bg-orange-100 rounded-lg">
          <Icon name="settings-3" className="text-orange-600 text-xl" />
        </div>
        <div>
          <h1 className="text-xl font-bold text-gray-800">Configuration</h1>
          <p className="text-sm text-gray-500 font-normal">Load raw config content by path, validate YAML before saving, and persist changes only when you click Save Config.</p>
        </div>
      </div>

      <div className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
        <div className="flex flex-col gap-4 border-b border-gray-100 bg-slate-50 px-4 py-4 lg:flex-row lg:items-center lg:justify-between">
          <div className="flex flex-1 flex-col gap-3">
            <Field
              label="Config file"
              hint={loadingPaths
                ? 'Discovering the known GoBackup config locations exposed by the backend.'
                : availablePaths.length > 0
                  ? missingPaths.length > 0
                    ? 'Existing config files are listed first. Missing allowlisted paths can be selected, but creation is only allowed after you confirm Save Config.'
                    : 'Choose one of the known existing config locations to load its raw YAML content.'
                  : 'The selector stays read-only until the backend exposes at least one allowlisted config path.'}
              size="small"
            >
              <div className="flex w-full items-center gap-3 lg:max-w-xl">
                <Select
                  value={selectedPath || ''}
                  onChange={handlePathChange}
                  disabled={loadingPaths || availablePaths.length === 0 || loadingContent || saving || validating}
                  appearance="outline"
                  className="w-full"
                >
                  {availablePaths.length === 0 ? (
                    <option value="">No config paths available</option>
                  ) : (
                    <>
                      {existingPaths.length > 0 && existingPaths.map((pathOption) => (
                        <option key={pathOption.path} value={pathOption.path}>
                          {pathOption.path}
                        </option>
                      ))}
                      {missingPaths.length > 0 && missingPaths.map((pathOption) => (
                        <option key={pathOption.path} value={pathOption.path}>
                          {pathOption.path} (create new)
                        </option>
                      ))}
                    </>
                  )}
                </Select>

                {(loadingPaths || loadingContent || saving || validating) && <Spinner size="tiny" labelPosition="below" />}
              </div>
            </Field>
          </div>

          <div className="flex items-center gap-2 self-start lg:self-auto">
            <div className="rounded-full border border-gray-200 bg-white px-3 py-1 text-xs font-medium text-gray-500">
              {loadingContent ? 'Loading content' : saving ? 'Saving changes' : validating ? 'Validating YAML' : dirty ? 'Unsaved changes' : 'All changes saved'}
            </div>
            <Button
              appearance="secondary"
              disabled={!selectedPath || loadingPaths || loadingContent || saving || validating}
              icon={validating ? <Spinner size="tiny" /> : undefined}
              onClick={handleValidate}
            >
              {validating ? 'Validating...' : 'Validate YAML'}
            </Button>
            <Button
              appearance="primary"
              disabled={!selectedPath || loadingPaths || loadingContent || saving || validating || !dirty}
              icon={saving ? <Spinner size="tiny" /> : undefined}
              onClick={handleSave}
            >
              {saving ? 'Saving...' : 'Save Config'}
            </Button>
          </div>
        </div>

        <div className="flex flex-col gap-4 p-4">
          <div className={`rounded-lg border px-4 py-3 text-sm ${messageClassName}`}>
            {message.text}
          </div>

          <div className="rounded-xl border border-dashed border-slate-200 bg-slate-50 p-3">
            <div className="mb-3 flex items-center justify-between gap-3">
              <div>
                <div className="text-sm font-medium text-gray-700">Editor surface</div>
                <div className="text-xs text-gray-400 mt-1">
                  Monaco mirrors the selected config file with YAML syntax highlighting while keeping the existing path, loading, dirty-state, and explicit save flow intact. Missing allowlisted targets are only created after explicit confirmation.
                </div>
              </div>
              <div className="flex items-center gap-2 text-xs text-gray-400">
                <Icon name="file-code" className="text-base" />
                <span>{content.length} chars</span>
              </div>
            </div>

            <div className="overflow-hidden rounded-lg border border-gray-200 bg-white">
              <Editor
                height="320px"
                language="yaml"
                path={selectedPath || 'gobackup.yml'}
                value={content}
                onChange={handleContentChange}
                theme="light"
                loading={
                  <div className="flex min-h-[320px] items-center justify-center text-sm text-gray-400">
                    Loading Monaco editor...
                  </div>
                }
                options={editorOptions}
              />
            </div>
          </div>

          <div className="flex flex-col gap-2 text-xs text-gray-400 sm:flex-row sm:items-center sm:justify-between">
            <span>Selected path: {selectedPath || 'Not selected yet'}</span>
            <span>Editor state: {saving ? 'saving' : validating ? 'validating' : dirty ? 'dirty' : 'synced'}</span>
          </div>
        </div>
      </div>
    </div>
  );
}
