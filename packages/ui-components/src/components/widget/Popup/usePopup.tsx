import { useState } from "react";
import { PopupLogic, UsePopupProps } from "./Popup.types";

export const usePopup = (props: UsePopupProps): PopupLogic => {
  // States
  const [isOpen, setIsOpen] = useState<boolean>(false);
  const [isVisible, setIsVisible] = useState<boolean>(false);

  // Functions
  const openPopup = () => {
    setIsOpen(true);
    setIsVisible(true);
  };

  const closePopup = () => {
    setIsVisible(false);
    setTimeout(() => {
      setIsOpen(false);
    }, 300);
  };

  return {
    isOpen,
    isVisible,
    openPopup,
    closePopup,
  };
};
