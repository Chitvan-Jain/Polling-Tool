import { useEffect, useRef, useState } from "react";

const BASE_URL = import.meta.env.VITE_API_BASE_URL;

export function useLiveResults(slug) {
  const [data, setData] = useState(null);
  const [connectionState, setConnectionState] = useState("connecting");
  const sourceRef = useRef(null);

  useEffect(() => {
    if (!slug) return;

    const source = new EventSource(`${BASE_URL}/api/p/${slug}/live`, {
      withCredentials: true,
    });
    sourceRef.current = source;

    source.addEventListener("results", (event) => {
      try {
        setData(JSON.parse(event.data));
        setConnectionState("connected");
      } catch {
        // ignore a malformed payload rather than crashing the whole page
      }
    });

    source.onopen = () => setConnectionState("connected");

    source.onerror = () => {
      if (source.readyState === EventSource.CONNECTING) {
        setConnectionState("reconnecting");
      }
    };

    return () => {
      source.close();
    };
  }, [slug]);

  return { data, connectionState };
}