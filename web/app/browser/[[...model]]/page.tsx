export function generateStaticParams() {
  return [{ model: ['__placeholder__'] }];
}

import BrowserClient from './browser-client';

export default function BrowserPage({ params }: { params: { model?: string[] } }) {
  return <BrowserClient model={params.model} />;
}
