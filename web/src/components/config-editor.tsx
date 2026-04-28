'use client';

import { Button } from '@fluentui/react-components';
import Editor, { loader, type Monaco, type OnMount } from '@monaco-editor/react';
import { useEffect, useMemo, useRef, useState } from 'react';
import * as monacoEditor from 'monaco-editor';
import type { editor } from 'monaco-editor';
import EditorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker';
import { configureMonacoYaml, type MonacoYaml } from 'monaco-yaml';
import configSchemaJson from '@/generated/config-schema.json';
import Icon from '@/components/icon';
import YamlWorker from '@/components/yaml.worker?worker';

const CONFIG_MODEL_URI = 'file:///gobackup.config.yml';
const CONFIG_SCHEMA_URI = 'gobackup://schemas/config-schema.json';
const configSchema = configSchemaJson as Record<string, unknown>;

type MonacoEnvironment = {
  getWorker: (workerId: string, label: string) => Worker;
};

type MonacoGlobalScope = typeof globalThis & {
  MonacoEnvironment?: MonacoEnvironment;
};

loader.config({ monaco: monacoEditor });

let monacoEnvironmentConfigured = false;
let yamlSupport: MonacoYaml | null = null;

function ensureMonacoEnvironment() {
  if (monacoEnvironmentConfigured || typeof window === 'undefined') {
    return;
  }

  const monacoGlobal = globalThis as MonacoGlobalScope;
  monacoGlobal.MonacoEnvironment = {
    getWorker(_workerId, label) {
      if (label === 'yaml') {
        return new YamlWorker();
      }

      return new EditorWorker();
    },
  };

  monacoEnvironmentConfigured = true;
}

function ensureYamlSupport(monaco: Monaco) {
  if (yamlSupport) {
    return;
  }

  yamlSupport = configureMonacoYaml(monaco, {
    completion: true,
    enableSchemaRequest: false,
    format: true,
    hover: true,
    validate: true,
    yamlVersion: '1.2',
    schemas: [
      {
        fileMatch: [CONFIG_MODEL_URI],
        schema: configSchema,
        uri: CONFIG_SCHEMA_URI,
      },
    ],
  });
}

export interface ConfigEditorProps {
  value: string;
  onChange: (value: string) => void;
  onSave?: () => void;
  disabled?: boolean;
  readOnly?: boolean;
  loading?: boolean;
  height?: string | number;
  onValidationChange?: (isValid: boolean, markers: editor.IMarker[]) => void;
}

export default function ConfigEditor({
  value,
  onChange,
  onSave,
  disabled = false,
  readOnly = false,
  loading = false,
  height = 440,
  onValidationChange,
}: ConfigEditorProps) {
  const saveActionRef = useRef({
    disabled,
    isValid: true,
    loading,
    onSave,
    readOnly,
  });
  const [markers, setMarkers] = useState<editor.IMarker[]>([]);

  const hasErrors = useMemo(
    () => markers.some((marker) => marker.severity === monacoEditor.MarkerSeverity.Error),
    [markers]
  );
  const isEditorReadOnly = disabled || readOnly || loading;
  const canSave = Boolean(onSave) && !isEditorReadOnly && !hasErrors;

  useEffect(() => {
    saveActionRef.current = {
      disabled,
      isValid: !hasErrors,
      loading,
      onSave,
      readOnly,
    };
  }, [disabled, hasErrors, loading, onSave, readOnly]);

  function publishValidation(nextMarkers: editor.IMarker[]) {
    setMarkers(nextMarkers);
    const isValid = nextMarkers.every(
      (marker) => marker.severity !== monacoEditor.MarkerSeverity.Error
    );
    onValidationChange?.(isValid, nextMarkers);
  }

  function triggerSave() {
    const { disabled: saveDisabled, isValid, loading: saveLoading, onSave: saveFn, readOnly: saveReadOnly } = saveActionRef.current;

    if (!saveFn || saveDisabled || saveReadOnly || saveLoading || !isValid) {
      return;
    }

    saveFn();
  }

  const handleEditorMount: OnMount = (editorInstance, monaco) => {
    editorInstance.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS, () => {
      triggerSave();
    });

    publishValidation(
      monaco.editor.getModelMarkers({ resource: monaco.Uri.parse(CONFIG_MODEL_URI) })
    );
  };

  return (
    <div className="rounded-2xl border border-gray-200 bg-white shadow-sm overflow-hidden">
      <div className="flex flex-col gap-3 border-b border-gray-100 bg-gradient-to-r from-orange-50 to-white px-4 py-3 md:flex-row md:items-center md:justify-between">
        <div className="min-w-0">
          <div className="flex items-center gap-2 text-sm font-semibold text-gray-800">
            <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-white text-orange-600 shadow-sm">
              <Icon name="file-code" />
            </div>
            <div>
              <div>Configuration YAML</div>
              <div className="text-xs font-normal text-gray-500">
                Inline schema validation for gobackup.yml
              </div>
            </div>
          </div>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          <div
            className={[
              'inline-flex items-center gap-2 rounded-full border px-3 py-1 text-xs font-medium',
              hasErrors
                ? 'border-red-200 bg-red-50 text-red-700'
                : 'border-green-200 bg-green-50 text-green-700',
            ].join(' ')}
          >
            <Icon name={hasErrors ? 'alert' : 'check'} />
            <span>
              {hasErrors ? `${markers.length} validation issue${markers.length === 1 ? '' : 's'}` : 'Schema looks good'}
            </span>
          </div>

          {onSave && (
            <Button
              appearance="primary"
              disabled={!canSave}
              icon={<Icon name={loading ? 'loader-4' : 'save'} className={loading ? 'animate-spin' : ''} />}
              onClick={triggerSave}
            >
              Save
            </Button>
          )}
        </div>
      </div>

      <div className="relative">
        {loading && (
          <div className="absolute inset-0 z-10 flex items-center justify-center bg-white/80 backdrop-blur-sm">
            <div className="flex items-center gap-2 rounded-full border border-orange-100 bg-white px-4 py-2 text-sm text-gray-600 shadow-sm">
              <Icon name="loader-4" className="animate-spin text-orange-600" />
              <span>Loading configuration editor…</span>
            </div>
          </div>
        )}

        <Editor
          beforeMount={(monaco) => {
            ensureMonacoEnvironment();
            ensureYamlSupport(monaco);
          }}
          defaultLanguage="yaml"
          height={height}
          language="yaml"
          loading={<div className="p-6 text-sm text-gray-500">Preparing editor…</div>}
          onChange={(nextValue) => {
            onChange(nextValue ?? '');
          }}
          onMount={handleEditorMount}
          onValidate={publishValidation}
          options={{
            automaticLayout: true,
            contextmenu: !disabled,
            domReadOnly: isEditorReadOnly,
            fontSize: 13,
            formatOnPaste: !isEditorReadOnly,
            formatOnType: !isEditorReadOnly,
            glyphMargin: false,
            guides: {
              indentation: true,
            },
            insertSpaces: true,
            lineNumbersMinChars: 3,
            minimap: {
              enabled: false,
            },
            padding: {
              bottom: 16,
              top: 16,
            },
            quickSuggestions: !isEditorReadOnly,
            readOnly: isEditorReadOnly,
            renderLineHighlight: 'all',
            scrollBeyondLastLine: false,
            smoothScrolling: true,
            tabSize: 2,
            wordWrap: 'on',
          }}
          path={CONFIG_MODEL_URI}
          theme="light"
          value={value}
        />
      </div>
    </div>
  );
}
