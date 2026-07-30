import { Card } from "@/components/layout";
import { PopupProps } from "./Popup.types";
import cn from "classnames";

const Popup = (props: PopupProps) => {
  return (
    <>
      {props.logic.isOpen && (
        <div className="popup">
          <Card className={cn("popup-card", props.className)}>
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
