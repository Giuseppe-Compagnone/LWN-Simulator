import { ReactNode } from "react";

export enum EntityDetailsSectionLayout {
  Full = "full",
  Half = "half",
  Third = "third",
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

export interface EntityDetailsMetric {
  label: string;
  value: ReactNode;
  detail?: ReactNode;
  icon?: string;
  tone?: EntityDetailTone;
  mono?: boolean;
  format?: EntityDetailValueFormat;
}

export interface EntityDetailsEmptyState {
  title: string;
  description: string;
  icon?: string;
}

export interface EntityDetailSection {
  id?: string;
  title: string;
  description?: string;
  icon?: string;
  items: Array<EntityDetailItem>;
  metrics?: Array<EntityDetailsMetric>;
  emptyState?: EntityDetailsEmptyState;
  layout?: EntityDetailsSectionLayout;
}

export interface EntityDetailsProps {
  sections: Array<EntityDetailSection>;
}
