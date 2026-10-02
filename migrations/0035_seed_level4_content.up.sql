-- Level 4 (Relationships): lessons, quizzes, questions, options.
INSERT INTO lessons (id, level_id, title, content, order_index, created_at) VALUES
    (
        '00000000-0000-0000-0000-000000003101',
        '00000000-0000-0000-0000-000000000004',
        'Understanding Attachment Styles',
        'Attachment theory describes patterns in how people relate to closeness and separation in relationships, shaped largely by early caregiving experiences. Broadly, these patterns are described as secure (comfortable with closeness and independence), anxious (craving closeness, fearing abandonment), avoidant (valuing independence, uneasy with closeness), and disorganized (a mix, often shaped by inconsistent or frightening caregiving).\n\nAttachment style isn''t a life sentence or a label to diagnose yourself or others with -- it''s a lens for noticing recurring patterns, like why certain relationship dynamics feel familiar even when they aren''t working well.\n\nUnderstanding your own tendencies can make it easier to recognize a reaction as "this is my attachment pattern showing up" rather than "this relationship is definitely doomed" -- a distinction that opens room for communication and change.',
        1,
        UTC_TIMESTAMP()
    ),
    (
        '00000000-0000-0000-0000-000000003102',
        '00000000-0000-0000-0000-000000000004',
        'Setting and Holding Boundaries',
        'A boundary is a clear statement of what you need in order to feel safe and respected in a relationship -- not a punishment, and not about controlling someone else''s behavior. A boundary describes what you will do, not a demand for what the other person must do: "I will leave the conversation if it becomes disrespectful" rather than "you have to stop being disrespectful."\n\nFor people with a history of trauma, setting boundaries can feel unsafe at first, especially if past relationships punished directness. It''s common to over-apologize or soften a boundary until it loses its meaning. Practicing boundaries in lower-stakes situations first can help build the skill before you need it somewhere that matters more.\n\nA boundary that is communicated and then not held isn''t really a boundary -- consistency, even in small ways, is what makes it trustworthy to both you and the other person.',
        2,
        UTC_TIMESTAMP()
    ),
    (
        '00000000-0000-0000-0000-000000003103',
        '00000000-0000-0000-0000-000000000004',
        'Rebuilding Trust After Trauma',
        'Trauma, especially trauma caused by another person, can make trust feel risky in any relationship afterward -- even ones that have given no reason for concern. This isn''t a character flaw; it''s a protective mechanism that generalized from a specific threat to threats in general.\n\nRebuilding trust usually happens gradually, through small and consistent experiences rather than one reassuring conversation. Noticing when someone''s actions match their words over time -- not just believing a single promise -- is often a more reliable gauge than how safe a situation initially feels.\n\nIt is also reasonable to extend trust at different paces in different relationships. Trusting slowly is not the same as being unable to trust -- it can be a sign that you are being thoughtful about something that matters.',
        3,
        UTC_TIMESTAMP()
    );

INSERT INTO quizzes (id, lesson_id, created_at) VALUES
    ('00000000-0000-0000-0000-000000003201', '00000000-0000-0000-0000-000000003101', UTC_TIMESTAMP()),
    ('00000000-0000-0000-0000-000000003202', '00000000-0000-0000-0000-000000003102', UTC_TIMESTAMP()),
    ('00000000-0000-0000-0000-000000003203', '00000000-0000-0000-0000-000000003103', UTC_TIMESTAMP());

INSERT INTO quiz_questions (id, quiz_id, question, order_index) VALUES
    ('00000000-0000-0000-0000-000000003301', '00000000-0000-0000-0000-000000003201', 'How is attachment style best used, according to this lesson?', 1),
    ('00000000-0000-0000-0000-000000003302', '00000000-0000-0000-0000-000000003202', 'Which of these is an example of a boundary, as described in this lesson?', 1),
    ('00000000-0000-0000-0000-000000003303', '00000000-0000-0000-0000-000000003203', 'What does this lesson say is often a more reliable gauge of trust than a single promise?', 1);

INSERT INTO quiz_options (id, question_id, text, is_correct, order_index) VALUES
    ('00000000-0000-0000-0000-000000003401', '00000000-0000-0000-0000-000000003301', 'As a lens for noticing recurring relationship patterns', TRUE, 1),
    ('00000000-0000-0000-0000-000000003402', '00000000-0000-0000-0000-000000003301', 'As a permanent diagnosis of someone''s character', FALSE, 2),
    ('00000000-0000-0000-0000-000000003403', '00000000-0000-0000-0000-000000003301', 'As a way to predict the future with certainty', FALSE, 3),
    ('00000000-0000-0000-0000-000000003404', '00000000-0000-0000-0000-000000003301', 'As a label only a professional should ever discuss', FALSE, 4),

    ('00000000-0000-0000-0000-000000003411', '00000000-0000-0000-0000-000000003302', '"You have to stop being disrespectful"', FALSE, 1),
    ('00000000-0000-0000-0000-000000003412', '00000000-0000-0000-0000-000000003302', '"I will leave the conversation if it becomes disrespectful"', TRUE, 2),
    ('00000000-0000-0000-0000-000000003413', '00000000-0000-0000-0000-000000003302', '"You always do this to me"', FALSE, 3),
    ('00000000-0000-0000-0000-000000003414', '00000000-0000-0000-0000-000000003302', '"I guess it''s fine, don''t worry about it"', FALSE, 4),

    ('00000000-0000-0000-0000-000000003421', '00000000-0000-0000-0000-000000003303', 'How safe the situation feels in the moment', FALSE, 1),
    ('00000000-0000-0000-0000-000000003422', '00000000-0000-0000-0000-000000003303', 'Whether someone''s actions match their words over time', TRUE, 2),
    ('00000000-0000-0000-0000-000000003423', '00000000-0000-0000-0000-000000003303', 'How confident the other person sounds', FALSE, 3),
    ('00000000-0000-0000-0000-000000003424', '00000000-0000-0000-0000-000000003303', 'How quickly trust is extended', FALSE, 4);
