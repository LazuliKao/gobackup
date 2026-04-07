import { LazyLog, ScrollFollow } from 'react-lazylog';
import { api } from '@/lib/api';

export default function LogView() {
  return (
    <div className="log-wrapper flex-1">
      <ScrollFollow
        startFollowing
        render={({ follow, onScroll }) => (
          <LazyLog
            extraLines={1}
            enableSearch={true}
            url={api.getLogStreamUrl()}
            stream
            follow={follow}
            onScroll={onScroll}
            fetchOptions={{ credentials: 'include' }}
          />
        )}
      />
    </div>
  );
}
