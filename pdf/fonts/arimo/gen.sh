r=$1

pyftsubset $r/Arimo-Regular.ttf \
  --unicodes='U+0020-007F,U+00A0-00FF,U+2009,U+2013,U+202F,U+20AC' \
  --layout-features='kern' \
     --output-file='Arimo-Invoice-Regular.ttf'

pyftsubset $r/Arimo-Italic.ttf \
  --unicodes='U+0020-007F,U+00A0-00FF,U+2009,U+2013,U+202F,U+20AC' \
  --layout-features='kern' \
     --output-file='Arimo-Invoice-Italic.ttf'
