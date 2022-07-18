# brainfuck
You can read more about the language itself [here](https://en.wikipedia.org/wiki/Brainfuck).

## This package
brainfuck provides an interpreter for `Brainfuck` language.

Despite the core idea of Brainfuck language its performance and shortness of interpreter were not the issue.
This interpreter is dedicated rather to show Golang approaches as it was written as an assignment task.

## Implementation notes

Here are points that explain an interpreter implementation:

1. **Memory data type.** 
   We're using generics to define memory data type. This gives user freedom to pick any type that fits the best

2. **Commands type.** 
   CmdType is byte. While we're not planning to give user a possibility to extend the language too much byte typ

3. **EOL.** 
   Interpreter ignores unknown commands, hence '\r' and '\n' don't affect code execution. While user can write s
   Again, while there's no some specific requirement it seems to be reasonable to keep implementation simple.

4. **Commands and Data pointers.** 
   CmdPtrType and DataPtrType are int. It could be changed if we have some requirement for the size of memory or
   an amount of Brainfuck commands that we're assuming to process.

5. **Commands caching in loops.** 
   One of requirements for this interpreter was reading commands in the flight and avoiding pre-loading commands
   During the loop we're repeating the commands that were already read. It seems reasonable to keep it and to re
   while iterating a loop.
   A map was picked as a cache container. It makes easier to manage it:
   - no need to extend memory moving along the loop body
   - no need to deal with commands order when caching it
     In case of performance requirements it may be implemented with slice that will make reading from cache faster

6. **Commands overloading.** 
   Custom commands handlers have access to public members – Data, DataPtr, CmdPtr, Input and Output.
   So custom command handler can read and write data to memory, setup where other commands will read/write,
   manage what the next command will be and read/write data from/to user.
