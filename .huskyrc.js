/*
 * Copyright (C) 2017-present Arctic Ice Studio <development@arcticicestudio.com>
 * Copyright (C) 2017-present Sven Greb <development@svengreb.de>
 *
 * Project:    snowsaw
 * Repository: https://github.com/arcticicestudio/snowsaw
 * License:    MIT
 */

/**
 * @file The husky configuration.
 * @author Arctic Ice Studio <development@arcticicestudio.com>
 * @author Sven Greb <development@svengreb.de>
 * @see https://github.com/typicode/husky
 */

module.exports = {
  hooks: {
    "pre-commit": "lint-staged"
  }
};
