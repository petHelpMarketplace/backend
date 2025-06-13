-- Insert services for dogs
INSERT INTO services (name, description, animal_category_id) VALUES 
    ('Вигул', 'Dog Walking - regular walks for your dog', (SELECT id FROM animal_categories WHERE name = 'Собаки')),
    ('Перетримка', 'Boarding - temporary care for your dog at a facility', (SELECT id FROM animal_categories WHERE name = 'Собаки')),
    ('Догляд вдома', 'House Sitting - care for your dog in your home', (SELECT id FROM animal_categories WHERE name = 'Собаки')),
    ('Грумінг', 'Grooming - professional grooming services for dogs', (SELECT id FROM animal_categories WHERE name = 'Собаки'));

-- Insert services for cats
INSERT INTO services (name, description, animal_category_id) VALUES 
    ('Перетримка', 'Boarding - temporary care for your cat at a facility', (SELECT id FROM animal_categories WHERE name = 'Коти')),
    ('Догляд вдома', 'House Sitting - care for your cat in your home', (SELECT id FROM animal_categories WHERE name = 'Коти')),
    ('Грумінг', 'Grooming - professional grooming services for cats', (SELECT id FROM animal_categories WHERE name = 'Коти')); 