'use client';

import { FluentProvider, Toaster } from '@fluentui/react-components';
import { gobackupLightTheme } from '../lib/theme';

export function ClientLayout({ children }: { children: React.ReactNode }) {
  return (
    <FluentProvider theme={gobackupLightTheme}>
      {children}
      <Toaster />
    </FluentProvider>
  );
}
