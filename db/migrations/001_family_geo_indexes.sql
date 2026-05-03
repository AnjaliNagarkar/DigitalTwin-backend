-- Geographic filters on FAMILY (DigitalTwin /houses API).
-- Apply on MySQL 8+ after backup. If an index already exists, skip that line or use a migration tool.
--
-- Expected columns: DISTRICT_ID, TALUKA_ID, VILLAGE_ID on table FAMILY (not "households").

CREATE INDEX idx_family_village_id ON FAMILY (VILLAGE_ID);
CREATE INDEX idx_family_taluka_id ON FAMILY (TALUKA_ID);
CREATE INDEX idx_family_district_id ON FAMILY (DISTRICT_ID);

-- Speeds member aggregation scoped by family (see handlers/houses.go scoped pop stats).
CREATE INDEX idx_family_member_family_id ON FAMILY_MEMBER (FAMILY_ID);
