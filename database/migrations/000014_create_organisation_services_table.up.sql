CREATE TABLE IF NOT EXISTS organisation_services (
    organisation_id INTEGER,
    service_id      INTEGER,
    PRIMARY KEY (organisation_id, service_id),
    CONSTRAINT fk_org_services_org
        FOREIGN KEY (organisation_id) REFERENCES organisations(id),
    CONSTRAINT fk_org_services_service
        FOREIGN KEY (service_id) REFERENCES services(id)
); 