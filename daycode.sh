#!/bin/bash

# Thiết lập biến môi trường để tránh lỗi ký GPG
export GPG_TTY=$(tty)

# Kiểm tra tham số message commit
if [ -z "$1" ]; then
    echo "Vui lòng nhập message commit."
    exit 1
fi

# Hiển thị trạng thái trước khi add
echo "Trạng thái trước khi add:"
git status

# Thêm tất cả thay đổi
git add .

# Hiển thị trạng thái sau khi add
echo "Trạng thái sau khi add:"
git status

# Commit có ký GPG với message truyền vào
git commit -S -m "$1"

# Lấy tên nhánh hiện tại
branch=$(git rev-parse --abbrev-ref HEAD)

# Đẩy lên nhánh hiện tại
git push origin "$branch"
