import { useEffect } from "react";
import { UseOutisideAlerterProps } from "./useOutsideAlerter.types";

const useOutsideAlerter = (props: UseOutisideAlerterProps) => {
  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (
        props.ref.current &&
        e.target instanceof Node &&
        !props.ref.current.contains(e.target)
      ) {
        props.onClickOutside();
      }
    }

    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [props, props.ref]);
};

export default useOutsideAlerter;
