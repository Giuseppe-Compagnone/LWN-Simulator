import { Card, CardLayout, CardType } from "../Card";
import { type ReactNode } from "react";
import {
  EntityDetailTone,
  EntityDetailValueFormat,
  EntityDetailsProps,
  EntityDetailsSectionLayout,
} from "./EntityDetails.types";

const formatEnumValue = (value: string) =>
  value
    .replace(/[-_]+/g, " ")
    .replace(/\s+/g, " ")
    .trim();

const formatDetailValue = (
  value: ReactNode,
  format?: EntityDetailValueFormat,
) => {
  if (format === EntityDetailValueFormat.Enum && typeof value === "string") {
    return formatEnumValue(value);
  }

  return value;
};

const EntityDetails = (props: EntityDetailsProps) => {
  return (
    <div className="entity-details">
      {props.sections.map((section) => (
        <Card
          className={`entity-details-section entity-details-section--${section.layout || EntityDetailsSectionLayout.Half}`}
          layout={CardLayout.Padded}
          type={CardType.Default}
          key={section.id || section.title}
        >
          <div className="entity-details-section__header">
            <div className="entity-details-section__heading">
              {section.icon && (
                <span className="entity-details-section__icon">
                  <span className="material-symbols-outlined">
                    {section.icon}
                  </span>
                </span>
              )}
              <div>
                <h2>{section.title}</h2>
                {section.description && <p>{section.description}</p>}
              </div>
            </div>
          </div>
          <dl>
            {section.items.map((item) => (
              <div className="entity-detail-row" key={item.label}>
                <dt>{item.label}</dt>
                <dd
                  className={`entity-detail-value entity-detail-value--${item.tone || EntityDetailTone.Default}${item.mono ? " entity-detail-value--mono" : ""}`}
                >
                  {formatDetailValue(item.value, item.format)}
                </dd>
              </div>
            ))}
          </dl>
        </Card>
      ))}
    </div>
  );
};

export default EntityDetails;
