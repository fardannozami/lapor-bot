import { useSyncExternalStore } from "react";
import { useColorScheme as useRNColorScheme } from "react-native";

/**
 * To support static rendering, this value needs to be re-calculated on the client side for web.
 * `useSyncExternalStore` returns the server snapshot ('light') during SSR/first render and the
 * client snapshot (the real color scheme) after hydration, avoiding a setState-in-effect.
 */
const emptySubscribe = () => () => {};

export function useColorScheme() {
  const hasHydrated = useSyncExternalStore(
    emptySubscribe,
    () => true, // client snapshot
    () => false, // server snapshot
  );

  const colorScheme = useRNColorScheme();

  if (hasHydrated) {
    return colorScheme;
  }

  return "light";
}
