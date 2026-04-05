'use client';

import { useSearchParams } from 'next/navigation';
import BrowserClient from './browser-client';

export default function BrowserPage() {
  const searchParams = useSearchParams();
  const model = searchParams.get('model');
  
  return <BrowserClient model={model ? [model] : undefined} />;
}
