touch ~/.gitcookies
chmod 0600 ~/.gitcookies

git config --global http.cookiefile ~/.gitcookies

tr , \\t <<\__END__ >>~/.gitcookies
.googlesource.com,TRUE,/,TRUE,2147483647,o,git-fdc.yelp.com=1/O2z1y-6Wk8FgASri-cK578xtTHvl0AD1_Z5hp2og1Nc
__END__
