/*
 * Copyright (C) 2017-present Arctic Ice Studio <development@arcticicestudio.com>
 * Copyright (C) 2017-present Sven Greb <development@svengreb.de>
 *
 * Project:    snowsaw
 * Repository: https://github.com/arcticicestudio/snowsaw
 * License:    MIT
 */

/**
 * @file The lint-staged configuration.
 * @author Arctic Ice Studio <development@arcticicestudio.com>
 * @author Sven Greb <development@svengreb.de>
 * @see https://github.com/okonet/lint-staged#configuration
 */

module.exports = {
  "*.{js,json,md,yml}": "prettier --list-different",
  "*.md": "remark --no-stdout"
};
