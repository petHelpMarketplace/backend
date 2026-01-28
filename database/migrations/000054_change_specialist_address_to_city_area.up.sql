-- Спочатку видаляємо старий constraint. 
-- Ім'я 'specialists_address_id_fkey' є стандартним для Postgres, 
-- якщо ви не задавали його вручну при створенні таблиці.
ALTER TABLE specialists 
DROP CONSTRAINT IF EXISTS specialists_address_id_fkey;

-- Додаємо новий constraint, що посилається на city_areas
ALTER TABLE specialists 
ADD CONSTRAINT specialists_address_id_fkey 
FOREIGN KEY (address_id) REFERENCES city_areas(id);
