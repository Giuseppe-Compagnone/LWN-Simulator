import { RefObject } from "react";

export interface UseOutisideAlerterProps<T extends HTMLElement = HTMLElement> {
  ref: RefObject<T | null>;
  onClickOutside: () => void;
}
