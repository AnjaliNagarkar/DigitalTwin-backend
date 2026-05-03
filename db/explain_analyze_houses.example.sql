-- Example: inspect /houses query plan after deploying CTE + indexes.
-- Replace filter values and run in mysql client against your schema.

EXPLAIN ANALYZE
WITH page_families AS (
  SELECT * FROM FAMILY f
  WHERE f.LATITUDE IS NOT NULL AND f.LONGITUDE IS NOT NULL
    AND f.LATITUDE != 0 AND f.LONGITUDE != 0
    AND f.VILLAGE_ID = 12345
  ORDER BY f.FAMILY_ID
  LIMIT 500 OFFSET 0
)
SELECT f.FAMILY_ID
FROM page_families f
LIMIT 5;
