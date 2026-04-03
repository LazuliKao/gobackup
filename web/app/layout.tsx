import './globals.css';
import { ClientLayout } from './client-layout';

export const metadata = {
  title: 'GoBackup',
  description: 'Backup management dashboard',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body>
        <ClientLayout>{children}</ClientLayout>
      </body>
    </html>
  );
}
