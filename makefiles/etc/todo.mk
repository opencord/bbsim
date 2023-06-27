## -----------------------------------------------------------------------
## Intent: Cast value into a boolean string.
## NOTE: : Careful with line comment placement, when octolthorp is indented
##         in a makefile return value will have whitespace appended.
## -----------------------------------------------------------------------
boolean = $(if $(strip $($(1))),false,true)# trailing whitespace is bad here

## -----------------------------------------------------------------------
## Intet: Negate input value for conditional use.
##   Return success when input value is null.
## Usage:
##   $(info ** defined[true]  = $(call not,value))
##   $(info ** defined[false] = $(call not,$(null)))
##   $(if $(call not,varname),$(error varname= is not set))
## -----------------------------------------------------------------------
not = $(strip \
  $(foreach true-false\
    ,$(info true-false=$(true-false))$(strip $(call boolean,$(1)))\
      $(subst true,$(null),$(true-false))\
  )\
)
