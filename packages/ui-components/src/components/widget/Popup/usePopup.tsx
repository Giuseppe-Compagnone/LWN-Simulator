import { useState } from "react";
import { PopupLogic, UsePopupProps } from "./Popup.types";

export const usePopup = (props: UsePopupProps): PopupLogic => {
  // States
  const [isOpen, setIsOpen] = useState<boolean>(false);

  // Functions
  const openPopup = () => {
    setIsOpen(true);
  };

  const closePopup = () => {
    setIsOpen(false);
  };

  return {
    isOpen,
    openPopup,
    closePopup,
  };
};
