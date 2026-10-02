-- Level 3 (Regulation): lessons, quizzes, questions, options.
INSERT INTO lessons (id, level_id, title, content, order_index, created_at) VALUES
    (
        '00000000-0000-0000-0000-000000002101',
        '00000000-0000-0000-0000-000000000003',
        'Grounding Techniques for the Present Moment',
        'Grounding is a set of simple techniques that help bring your attention back to the present moment when you feel overwhelmed, anxious, or disconnected. The idea is to give your nervous system clear, concrete evidence that you are here, now, and safe -- rather than caught in a memory or a spiral of "what if."\n\nOne widely used technique is "5-4-3-2-1": name five things you can see, four things you can touch, three things you can hear, two things you can smell, and one thing you can taste. Another is simply pressing your feet into the floor and noticing the physical sensation of contact and support.\n\nGrounding doesn''t erase difficult feelings, and it isn''t meant to. It creates a small window of stability so you can respond to a situation rather than being swept along by it. Like any skill, it works better with practice during calm moments, not just emergencies.',
        1,
        UTC_TIMESTAMP()
    ),
    (
        '00000000-0000-0000-0000-000000002102',
        '00000000-0000-0000-0000-000000000003',
        'Sleep and the Nervous System',
        'Sleep and trauma recovery are closely linked. A nervous system that feels unsafe often struggles to fully power down, leading to lighter, more fragmented sleep, frequent waking, or vivid distressing dreams -- which in turn makes the nervous system more reactive the next day. It can become a difficult cycle.\n\nSome practical anchors help signal safety to your body at bedtime: a consistent wind-down routine, dimming lights an hour before bed, and avoiding reviewing stressful topics right before sleep. None of these are a cure on their own, but consistency matters more than any single technique.\n\nIf sleep difficulties are severe or ongoing, that is worth raising with a healthcare provider -- this lesson is general education, not a substitute for a clinical sleep assessment.',
        2,
        UTC_TIMESTAMP()
    ),
    (
        '00000000-0000-0000-0000-000000002103',
        '00000000-0000-0000-0000-000000000003',
        'Everyday Stress Management',
        'Regulation isn''t only for moments of crisis -- it''s also the everyday habits that keep your baseline stress from building up unnoticed until something small tips it over. Regular movement, even a short walk, helps metabolize stress hormones your body produced but never got to use the way it would after, say, running from danger.\n\nBreath is another accessible tool: slow, extended exhales (longer than your inhale) activate your body''s calming response. A common pattern is inhaling for four counts and exhaling for six, repeated for a minute or two.\n\nThe goal isn''t to eliminate stress -- some stress is a normal part of life. It''s to build a few reliable, low-effort tools you can reach for before stress accumulates past what feels manageable.',
        3,
        UTC_TIMESTAMP()
    );

INSERT INTO quizzes (id, lesson_id, created_at) VALUES
    ('00000000-0000-0000-0000-000000002201', '00000000-0000-0000-0000-000000002101', UTC_TIMESTAMP()),
    ('00000000-0000-0000-0000-000000002202', '00000000-0000-0000-0000-000000002102', UTC_TIMESTAMP()),
    ('00000000-0000-0000-0000-000000002203', '00000000-0000-0000-0000-000000002103', UTC_TIMESTAMP());

INSERT INTO quiz_questions (id, quiz_id, question, order_index) VALUES
    ('00000000-0000-0000-0000-000000002301', '00000000-0000-0000-0000-000000002201', 'What is the purpose of a grounding technique like "5-4-3-2-1"?', 1),
    ('00000000-0000-0000-0000-000000002302', '00000000-0000-0000-0000-000000002202', 'Why can trauma make sleep more difficult?', 1),
    ('00000000-0000-0000-0000-000000002303', '00000000-0000-0000-0000-000000002203', 'In the breathing pattern described, which part of the breath is extended?', 1);

INSERT INTO quiz_options (id, question_id, text, is_correct, order_index) VALUES
    ('00000000-0000-0000-0000-000000002401', '00000000-0000-0000-0000-000000002301', 'To bring attention back to the present moment', TRUE, 1),
    ('00000000-0000-0000-0000-000000002402', '00000000-0000-0000-0000-000000002301', 'To permanently erase difficult feelings', FALSE, 2),
    ('00000000-0000-0000-0000-000000002403', '00000000-0000-0000-0000-000000002301', 'To help you fall asleep faster', FALSE, 3),
    ('00000000-0000-0000-0000-000000002404', '00000000-0000-0000-0000-000000002301', 'To replace professional treatment', FALSE, 4),

    ('00000000-0000-0000-0000-000000002411', '00000000-0000-0000-0000-000000002302', 'A nervous system that feels unsafe struggles to fully power down', TRUE, 1),
    ('00000000-0000-0000-0000-000000002412', '00000000-0000-0000-0000-000000002302', 'Trauma has no effect on sleep', FALSE, 2),
    ('00000000-0000-0000-0000-000000002413', '00000000-0000-0000-0000-000000002302', 'It only affects how long you sleep, not quality', FALSE, 3),
    ('00000000-0000-0000-0000-000000002414', '00000000-0000-0000-0000-000000002302', 'It makes you need less sleep overall', FALSE, 4),

    ('00000000-0000-0000-0000-000000002421', '00000000-0000-0000-0000-000000002303', 'The inhale', FALSE, 1),
    ('00000000-0000-0000-0000-000000002422', '00000000-0000-0000-0000-000000002303', 'The exhale', TRUE, 2),
    ('00000000-0000-0000-0000-000000002423', '00000000-0000-0000-0000-000000002303', 'Both equally', FALSE, 3),
    ('00000000-0000-0000-0000-000000002424', '00000000-0000-0000-0000-000000002303', 'Neither -- the pattern uses breath holds only', FALSE, 4);
