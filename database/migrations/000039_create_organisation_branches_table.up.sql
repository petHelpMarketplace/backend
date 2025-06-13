CREATE TABLE IF NOT EXISTS organisation_branches (
    organisation_id INTEGER,
    branch_id       INTEGER,
    PRIMARY KEY (organisation_id, branch_id),
    CONSTRAINT fk_org_branches_org
        FOREIGN KEY (organisation_id) REFERENCES organisations(id),
    CONSTRAINT fk_org_branches_branch
        FOREIGN KEY (branch_id) REFERENCES branches(id)
); 