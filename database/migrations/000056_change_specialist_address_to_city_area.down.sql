-- Видаляємо зв'язок з city_areas
ALTER TABLE specialists 
DROP CONSTRAINT IF EXISTS specialists_address_id_fkey;

-- Додаємо новий constraint, що посилається на addresses

-- Повертаємо зв'язок з addresses
ALTER TABLE specialists 
ADD CONSTRAINT specialists_address_id_fkey 
FOREIGN KEY (address_id) REFERENCES addresses(id);
