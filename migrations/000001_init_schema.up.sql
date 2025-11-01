CREATE TABLE role (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name varchar(50) NOT NULL UNIQUE,
  is_deleted boolean DEFAULT false,
  updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE coffee_shop (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name varchar(100) NOT NULL,
  address varchar(255) NOT NULL,
  contacts varchar(100),
  welcome_message text,
  rules text,
  updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE status (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  title varchar(50) NOT NULL UNIQUE,
  is_deleted boolean DEFAULT false,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name varchar(100) NOT NULL,
  phone varchar(15) NOT NULL UNIQUE,
  role_id uuid REFERENCES role(id),
  is_deleted boolean DEFAULT false,
  updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE category (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  coffee_shop_id uuid REFERENCES coffee_shop(id),
  title varchar(50) NOT NULL,
  description text,
  is_deleted boolean DEFAULT false,
  updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE reward_type (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  coffee_shop_id uuid REFERENCES coffee_shop(id),
  description text NOT NULL,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE idea (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid REFERENCES users(id),
  coffee_shop_id uuid REFERENCES coffee_shop(id),
  category_id uuid REFERENCES category(id),
  status_id uuid REFERENCES status(id),
  title varchar(150) NOT NULL,
  description text NOT NULL,
  image_url varchar(255),
  is_deleted boolean DEFAULT false,
  updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE idea_like(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid REFERENCES users(id),
  idea_id uuid REFERENCES idea(id),
  created_at timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE reward(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  receiver_id uuid REFERENCES users(id),
  coffee_shop_id uuid REFERENCES coffee_shop(id),
  idea_id uuid REFERENCES idea(id),
  reward_type_id uuid REFERENCES reward_type(id),
  is_activated boolean DEFAULT false,
  given_at timestamp,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE idea_comment (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  creator_id uuid REFERENCES users(id),
  idea_id uuid REFERENCES idea(id),
  text text NOT NULL,
  is_deleted boolean DEFAULT false,
  updated_at timestamp DEFAULT CURRENT_TIMESTAMP,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE worker_coffee_shop (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  worker_id uuid REFERENCES users(id),
  coffee_shop_id uuid REFERENCES coffee_shop(id),
  is_deleted boolean DEFAULT false,
  created_at timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE banned_user (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid REFERENCES users(id),
  coffee_shop_id uuid REFERENCES coffee_shop(id),
  created_at timestamp DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_phone ON users(phone);
CREATE INDEX idx_user_role_id ON users(role_id);
CREATE INDEX idx_idea_coffee_shop_id ON idea(coffee_shop_id);
CREATE INDEX idx_idea_user_id ON idea(user_id);
CREATE INDEX idx_idea_status_id ON idea(status_id);
CREATE INDEX idx_idea_category_id ON idea(category_id);
CREATE INDEX idx_idea_created_at ON idea(created_at DESC);
CREATE INDEX idx_idea_coffee_shop_status ON idea(coffee_shop_id, status_id);
CREATE INDEX idx_idea_coffee_shop_category ON idea(coffee_shop_id, category_id);
CREATE INDEX idx_like_idea_id ON idea_like(idea_id);
CREATE INDEX idx_like_user_idea ON idea_like(user_id, idea_id);
CREATE INDEX idx_reward_receiver_id ON reward(receiver_id);
CREATE INDEX idx_reward_coffee_shop_id ON reward(coffee_shop_id);
CREATE INDEX idx_reward_idea_id ON reward(idea_id);
CREATE INDEX idx_comment_idea_id ON idea_comment(idea_id);
CREATE INDEX idx_comment_creator_id ON idea_comment(creator_id);
CREATE INDEX idx_category_coffee_shop_id ON category(coffee_shop_id);
CREATE INDEX idx_worker_coffee_shop_user ON worker_coffee_shop(worker_id, coffee_shop_id);
CREATE INDEX idx_banned_user_check ON banned_user(user_id, coffee_shop_id);
