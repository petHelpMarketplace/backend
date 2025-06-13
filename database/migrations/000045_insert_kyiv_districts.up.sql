-- Insert Kyiv districts
INSERT INTO city_areas (id, city_id, area_name) VALUES 
    (1, (SELECT id FROM cities WHERE city = 'Київ'), 'Голосіївський район'),
    (2, (SELECT id FROM cities WHERE city = 'Київ'), 'Дарницький район'),
    (3, (SELECT id FROM cities WHERE city = 'Київ'), 'Деснянський район'),
    (4, (SELECT id FROM cities WHERE city = 'Київ'), 'Дніпровський район'),
    (5, (SELECT id FROM cities WHERE city = 'Київ'), 'Оболонський район'),
    (6, (SELECT id FROM cities WHERE city = 'Київ'), 'Печерський район'),
    (7, (SELECT id FROM cities WHERE city = 'Київ'), 'Подільський район'),
    (8, (SELECT id FROM cities WHERE city = 'Київ'), 'Святошинський район'),
    (9, (SELECT id FROM cities WHERE city = 'Київ'), 'Солом''янський район'),
    (10, (SELECT id FROM cities WHERE city = 'Київ'), 'Шевченківський район'); 