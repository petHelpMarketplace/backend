-- Remove services for dogs and cats
DELETE FROM services 
WHERE name IN ('Вигул', 'Перетримка', 'Догляд вдома', 'Грумінг')
AND animal_category_id IN (
    (SELECT id FROM animal_categories WHERE name = 'Собаки'),
    (SELECT id FROM animal_categories WHERE name = 'Коти')
); 