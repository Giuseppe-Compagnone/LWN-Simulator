import { ReactNode } from "react";

export enum EntityDetailsSectionLayout {
  Full = "full",
  Half = "half",
}

export enum EntityDetailTone {
  Default = "default",
  Positive = "positive",
  Warning = "warning",
  Negative = "negative",
  Muted = "muted",
}

export enum EntityDetailValueFormat {
  Default = "default",
  Enum = "enum",
}

export interface EntityDetailItem {
  label: string;
  value: ReactNode;
  tone?: EntityDetailTone;
  mono?: boolean;
  format?: EntityDetailValueFormat;
}

export interface EntityDetailSection {
  id?: string;
  title: string;
  description?: string;
  icon?: string;
  items: Array<EntityDetailItem>;
  layout?: EntityDetailsSectionLayout;
}

export interface EntityDetailsProps {
  sections: Array<EntityDetailSection>;
}
