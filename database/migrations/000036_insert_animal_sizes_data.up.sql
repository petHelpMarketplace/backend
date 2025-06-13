-- Dogs sizes
INSERT INTO animal_sizes (animal_category_id, name, min_weight, max_weight) VALUES
(1, 'Small', 0, 7),
(1, 'Medium', 7, 18),
(1, 'Large', 18, 45),
(1, 'Extra Large', 45, NULL);

-- Cats sizes
INSERT INTO animal_sizes (animal_category_id, name, min_weight, max_weight) VALUES
(2, 'Small', 0, 3),
(2, 'Medium', 3, 5),
(2, 'Large', 5, 8),
(2, 'Extra Large', 8, NULL);

-- Rodents sizes
INSERT INTO animal_sizes (animal_category_id, name, min_weight, max_weight) VALUES
(3, 'Small', 0, 0.5),
(3, 'Medium', 0.5, 2),
(3, 'Large', 2, NULL);

-- Birds sizes
INSERT INTO animal_sizes (animal_category_id, name, min_weight, max_weight) VALUES
(4, 'Small', 0, 0.3),
(4, 'Medium', 0.3, 1),
(4, 'Large', 1, NULL);

-- Reptiles sizes
INSERT INTO animal_sizes (animal_category_id, name, min_weight, max_weight) VALUES
(5, 'Small', 0, 1),
(5, 'Medium', 1, 5),
(5, 'Large', 5, NULL); 