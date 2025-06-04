alter table PortalProp
add column zone VARCHAR(256) NOT NULL;

alter table PublishedProperty
add column plan ENUM("simple", "highlighted", "super") DEFAULT "simple" NOT NULL;
