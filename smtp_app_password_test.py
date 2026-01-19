import os
import smtplib
import ssl
import sys

# Simple SMTP App Password test for Gmail (or any STARTTLS server)
# Usage:
#   SMTP_HOST=smtp.gmail.com \
#   SMTP_PORT=587 \
#   SMTP_USER=your.name@gmail.com \
#   SMTP_PASS=your_app_password \
#   SMTP_TO=recipient@example.com \
#   python3 smtp_app_password_test.py


def getenv(key: str, default: str = "") -> str:
    return os.getenv(key, default)


def main() -> int:
    host = getenv("SMTP_HOST", "smtp.gmail.com")
    port = int(getenv("SMTP_PORT", "587"))
    user = getenv("SMTP_USER","developertyrone@gmail.com")
    password = getenv("SMTP_PASS", "gzcw ejfr vsft yluk ")
    to_addr = getenv("SMTP_TO", "temp.superkufu@gmail.com")

    if not user or not password or not to_addr:
        print("Missing required envs: SMTP_USER, SMTP_PASS, SMTP_TO", file=sys.stderr)
        return 1

    msg = f"""From: {user}\r\nTo: {to_addr}\r\nSubject: SMTP App Password Test\r\n\r\nThis is a test email sent using an app password.\r\n"""

    context = ssl.create_default_context()

    try:
        with smtplib.SMTP(host, port, timeout=20) as server:
            server.ehlo()
            server.starttls(context=context)
            server.ehlo()
            server.login(user, password)
            server.sendmail(user, [to_addr], msg)
        print(f"✅ SMTP send succeeded to {to_addr}")
        return 0
    except Exception as e:
        print(f"❌ SMTP send failed: {e}", file=sys.stderr)
        return 2


if __name__ == "__main__":
    sys.exit(main())
