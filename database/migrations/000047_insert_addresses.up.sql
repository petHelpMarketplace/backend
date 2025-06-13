INSERT INTO countries (id, name) VALUES (1, 'Україна') ON CONFLICT (id) DO NOTHING;

INSERT INTO cities (id, country_id, city) VALUES (1, 1, 'Київ') ON CONFLICT (id) DO NOTHING;

INSERT INTO addresses (
    id,
    country_id,
    city_id,
    area_id,
    postal_code,
    street,
    apt
) VALUES 
    (1, (SELECT id FROM countries WHERE name = 'Україна'),
        (SELECT id FROM cities WHERE city = 'Київ'),
        (SELECT id FROM city_areas WHERE area_name = 'Голосіївський район'),
        '02152', 'вул. Березняківська', '12'),
        
    (2, (SELECT id FROM countries WHERE name = 'Україна'),
        (SELECT id FROM cities WHERE city = 'Київ'),
        (SELECT id FROM city_areas WHERE area_name = 'Дарницький район'),
        '02068', 'вул. Драгоманова', '31'),
        
    (3, (SELECT id FROM countries WHERE name = 'Україна'),
        (SELECT id FROM cities WHERE city = 'Київ'),
        (SELECT id FROM city_areas WHERE area_name = 'Дарницький район'),
        '02095', 'вул. Тростянецька', '12'),
        
    (4, (SELECT id FROM countries WHERE name = 'Україна'),
        (SELECT id FROM cities WHERE city = 'Київ'),
        (SELECT id FROM city_areas WHERE area_name = 'Печерський район'),
        '01015', 'вул. Лаврська', '44'),
        
    (5, (SELECT id FROM countries WHERE name = 'Україна'),
        (SELECT id FROM cities WHERE city = 'Київ'),
        (SELECT id FROM city_areas WHERE area_name = 'Оболонський район'),
        '04210', 'просп. Оболонський', '57'),
        
    (6, (SELECT id FROM countries WHERE name = 'Україна'),
        (SELECT id FROM cities WHERE city = 'Київ'),
        (SELECT id FROM city_areas WHERE area_name = 'Дарницький район'),
        '02140', 'вул. Бориса Гмирі', '34'),
        
    (7, (SELECT id FROM countries WHERE name = 'Україна'),
        (SELECT id FROM cities WHERE city = 'Київ'),
        (SELECT id FROM city_areas WHERE area_name = 'Святошинський район'),
        '03148', 'вул. Якуба Коласа', '10'),
        
    (8, (SELECT id FROM countries WHERE name = 'Україна'),
        (SELECT id FROM cities WHERE city = 'Київ'),
        (SELECT id FROM city_areas WHERE area_name = 'Шевченківський район'),
        '01133', 'вул. Євгена Коновальця', '11'),
        
    (9, (SELECT id FROM countries WHERE name = 'Україна'),
        (SELECT id FROM cities WHERE city = 'Київ'),
        (SELECT id FROM city_areas WHERE area_name = 'Печерський район'),
        '01103', 'вул. Бастіонна', '3'),
        
    (10, (SELECT id FROM countries WHERE name = 'Україна'),
         (SELECT id FROM cities WHERE city = 'Київ'),
         (SELECT id FROM city_areas WHERE area_name = 'Шевченківський район'),
         '04116', 'вул. Златоустівська', '78'),
        
    (11, (SELECT id FROM countries WHERE name = 'Україна'),
         (SELECT id FROM cities WHERE city = 'Київ'),
         (SELECT id FROM city_areas WHERE area_name = 'Деснянський район'),
         '02225', 'вул. Бальзака', '114'),
        
    (12, (SELECT id FROM countries WHERE name = 'Україна'),
         (SELECT id FROM cities WHERE city = 'Київ'),
         (SELECT id FROM city_areas WHERE area_name = 'Голосіївський район'),
         '02152', 'вул. Березняківська', '57'),
        
    (13, (SELECT id FROM countries WHERE name = 'Україна'),
         (SELECT id FROM cities WHERE city = 'Київ'),
         (SELECT id FROM city_areas WHERE area_name = 'Святошинський район'),
         '01133', 'вул. Лесі Українки', '67'),
        
    (14, (SELECT id FROM countries WHERE name = 'Україна'),
         (SELECT id FROM cities WHERE city = 'Київ'),
         (SELECT id FROM city_areas WHERE area_name = 'Оболонський район'),
         '04210', 'вул. Йорданська', '45'),
        
    (15, (SELECT id FROM countries WHERE name = 'Україна'),
         (SELECT id FROM cities WHERE city = 'Київ'),
         (SELECT id FROM city_areas WHERE area_name = 'Голосіївський район'),
         '02147', 'вул. Серафимовича', '23'),
        
    (16, (SELECT id FROM countries WHERE name = 'Україна'),
         (SELECT id FROM cities WHERE city = 'Київ'),
         (SELECT id FROM city_areas WHERE area_name = 'Дарницький район'),
         '02095', 'вул. Княжий Затон', '1')
ON CONFLICT (id) DO NOTHING;

