import { render } from 'preact';
import { FluentProvider, Toaster } from '@fluentui/react-components';
import { gobackupLightTheme } from '@/lib/theme';
import Home from '@/app/page';
import '@/app/globals.css';

render(
  <FluentProvider theme={gobackupLightTheme}>
    <Home />
    <Toaster />
  </FluentProvider>,
  document.getElementById('root')!
);
