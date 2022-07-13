Program find the README.md file in current directory.  

If `doc` directory exists, confugo will find all *.md files inside it and create a new document for every md file 
in confluence document tree, as sibling of ID mentioned in `confluence.ancestor` file.

Confugo can be part of CI process.