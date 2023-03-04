chromedriver proxy for wsl
----

This is chromedriver proxy for wsl.
You can use Chrome-on-Windows from wsl like a local webdriver.

## Getting started

1. Install Chrome & ChromeDriver on Windows.
   ref. https://chromedriver.chromium.org/downloads
2. Put binary on Ubuntu/WSL.
3. Add `chromedriver_wsl_config.json` to same directory.```
    ```json
   {
      "chromedriver_bin": "/mnt/c/chromedriver_win32/chromedriver.exe"
   }
    ```
   Change path your installed direcotry.
4. Run this binary like chromedriver.
