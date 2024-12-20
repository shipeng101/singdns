import { useEffect, useRef } from 'react';

function usePolling(callback, interval = 5000, enabled = true) {
  const savedCallback = useRef();
  const timeoutRef = useRef();

  useEffect(() => {
    savedCallback.current = callback;
  }, [callback]);

  useEffect(() => {
    function tick() {
      if (enabled) {
        savedCallback.current();
      }
    }

    if (enabled) {
      const id = setInterval(tick, interval);
      timeoutRef.current = id;
      return () => clearInterval(id);
    }

    return () => {
      if (timeoutRef.current) {
        clearInterval(timeoutRef.current);
      }
    };
  }, [interval, enabled]);
}

export default usePolling; 