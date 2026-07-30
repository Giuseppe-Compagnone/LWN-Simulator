import { Card } from "@/components/layout";
import { PopupProps } from "./Popup.types";
import cn from "classnames";
import { useRef } from "react";
import { useOutsideAlerter } from "@/hooks";

const Popup = (props: PopupProps) => {
  // Refs
  const cardRef = useRef<HTMLDivElement>(null);

  useOutsideAlerter({
    ref: cardRef,
    onClickOutside: () => {
      console.log("CLICK");
      props.logic.closePopup();
    },
  });

  return (
    <>
      {props.logic.isOpen && (
        <div
          className={cn(
            "popup",
            props.logic.isVisible ? "visible" : "not-visible",
          )}
        >
          <Card className={cn("popup-card", props.className)} ref={cardRef}>
            <div className="popup-header">
              <strong>{props.title}</strong>
              <span
                className="material-symbols-outlined"
                onClick={() => {
                  props.logic.closePopup();
                }}
              >
                close
              </span>
            </div>
            <div className="popup-body">{props.children}</div>
          </Card>
        </div>
      )}
    </>
  );
};

export default Popup;
