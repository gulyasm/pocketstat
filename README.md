# Pocket stat library

## Install
`go install`

## Usage

The config file has to be set via the `config` flag.

Three different format can be specified with the `format` flag.
    
   - `csv`
   - `db`
   - `console`

If you specified csv, than two verbosity level can be set by the `v` flag.
   
   - `elements`
   - `count`

### Example
`./pocket-stat -config="/home/gulyasm/.pocket-stat" -format="csv" -v="elements"`

For further details, read the code.
