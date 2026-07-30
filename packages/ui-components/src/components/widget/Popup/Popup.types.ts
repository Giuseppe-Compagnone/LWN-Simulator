import { PropsWithChildren } from "react";

export interface PopupProps extends PropsWithChildren {
  className?: string;
  logic: PopupLogic;
  title: string;
}

export interface UsePopupProps {}

export interface PopupLogic {
  isOpen: boolean;
  openPopup: () => void;
  closePopup: () => void;
}
