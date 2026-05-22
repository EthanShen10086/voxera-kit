import type { AIClient } from './client';
import type { Message, TokenUsage } from './types';

export interface ChatState {
  messages: Message[];
  isLoading: boolean;
  error: Error | null;
  usage?: TokenUsage;
}

export function createChatHook(
  client: AIClient,
  useState: <T>(init: T) => [T, (v: T | ((prev: T) => T)) => void],
  useEffect: (fn: () => void | (() => void), deps: unknown[]) => void,
  useCallback: <T extends (...args: any[]) => any>(fn: T, deps: unknown[]) => T,
) {
  return function useChat(options?: {
    model?: string;
    systemPrompt?: string;
  }): {
    messages: Message[];
    isLoading: boolean;
    error: Error | null;
    send: (content: string) => void;
    abort: () => void;
    reset: () => void;
  } {
    const initialMessages: Message[] = options?.systemPrompt
      ? [{ role: 'system' as const, content: options.systemPrompt }]
      : [];

    const [messages, setMessages] = useState<Message[]>(initialMessages);
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<Error | null>(null);

    useEffect(() => {
      return () => {
        client.abort();
      };
    }, []);

    const send = useCallback(
      ((content: string) => {
        const userMessage: Message = {
          role: 'user',
          content,
          timestamp: Date.now(),
        };

        setMessages((prev: Message[]) => [...prev, userMessage]);
        setIsLoading(true);
        setError(null);

        const assistantMessage: Message = {
          role: 'assistant',
          content: '',
          timestamp: Date.now(),
        };
        setMessages((prev: Message[]) => [...prev, assistantMessage]);

        let fullContent = '';

        client
          .chatStream(
            {
              model: options?.model,
              messages: [...messages, userMessage],
              stream: true,
            },
            (chunk) => {
              if (chunk.done) return;
              fullContent += chunk.content;
              setMessages((prev: Message[]) => {
                const updated = [...prev];
                updated[updated.length - 1] = {
                  ...updated[updated.length - 1]!,
                  content: fullContent,
                };
                return updated;
              });
            },
          )
          .then(() => {
            setIsLoading(false);
          })
          .catch((err: unknown) => {
            const e = err instanceof Error ? err : new Error(String(err));
            setError(e);
            setIsLoading(false);
            setMessages((prev: Message[]) => prev.slice(0, -1));
          });
      }) as (content: string) => void,
      [messages, options?.model],
    );

    const abort = useCallback(
      (() => {
        client.abort();
        setIsLoading(false);
      }) as () => void,
      [],
    );

    const reset = useCallback(
      (() => {
        client.abort();
        setMessages(initialMessages);
        setError(null);
        setIsLoading(false);
      }) as () => void,
      [],
    );

    return { messages, isLoading, error, send, abort, reset };
  };
}

export function createCompletionHook(
  client: AIClient,
  useState: <T>(init: T) => [T, (v: T | ((prev: T) => T)) => void],
  useCallback: <T extends (...args: any[]) => any>(fn: T, deps: unknown[]) => T,
) {
  return function useCompletion(): {
    text: string;
    isLoading: boolean;
    error: Error | null;
    complete: (prompt: string, options?: { model?: string }) => void;
    abort: () => void;
  } {
    const [text, setText] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<Error | null>(null);

    const complete = useCallback(
      ((prompt: string, options?: { model?: string }) => {
        setIsLoading(true);
        setError(null);
        setText('');

        let fullContent = '';

        client
          .chatStream(
            {
              model: options?.model,
              messages: [{ role: 'user', content: prompt }],
              stream: true,
            },
            (chunk) => {
              if (chunk.done) return;
              fullContent += chunk.content;
              setText(fullContent);
            },
          )
          .then(() => {
            setIsLoading(false);
          })
          .catch((err: unknown) => {
            const e = err instanceof Error ? err : new Error(String(err));
            setError(e);
            setIsLoading(false);
          });
      }) as (prompt: string, options?: { model?: string }) => void,
      [],
    );

    const abort = useCallback(
      (() => {
        client.abort();
        setIsLoading(false);
      }) as () => void,
      [],
    );

    return { text, isLoading, error, complete, abort };
  };
}

export function createSummaryHook(
  client: AIClient,
  useState: <T>(init: T) => [T, (v: T | ((prev: T) => T)) => void],
  useCallback: <T extends (...args: any[]) => any>(fn: T, deps: unknown[]) => T,
) {
  return function useSummary(): {
    summary: string;
    isLoading: boolean;
    error: Error | null;
    summarize: (text: string) => void;
  } {
    const [summary, setSummary] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<Error | null>(null);

    const summarize = useCallback(
      ((text: string) => {
        setIsLoading(true);
        setError(null);
        setSummary('');

        client
          .summarize(text)
          .then((result) => {
            setSummary(result);
            setIsLoading(false);
          })
          .catch((err: unknown) => {
            const e = err instanceof Error ? err : new Error(String(err));
            setError(e);
            setIsLoading(false);
          });
      }) as (text: string) => void,
      [],
    );

    return { summary, isLoading, error, summarize };
  };
}
