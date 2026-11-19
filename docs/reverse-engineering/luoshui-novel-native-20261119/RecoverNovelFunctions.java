// Read-only fallback: decompile selected functions from the immutable snapshot.
// This is analyst-run Ghidra output, not an REA Evidence record.
import ghidra.app.util.headless.HeadlessScript;
import ghidra.app.decompiler.*;
import ghidra.program.model.listing.*;
import java.nio.file.*;
import java.nio.charset.StandardCharsets;

public class RecoverNovelFunctions extends HeadlessScript {
 public void run() throws Exception {
  String[] args=getScriptArgs();Path output=Paths.get(args[0]);Files.createDirectories(output);
  DecompInterface decomp=new DecompInterface();decomp.openProgram(currentProgram);
  for(int i=1;i<args.length;i++){
   Function f=getFunctionAt(toAddr(args[i]));
   if(f==null){println("UNRESOLVED "+args[i]);continue;}
   DecompileResults result=decomp.decompileFunction(f,90,monitor);
   if(result.decompileCompleted()){
    String text="/* Analyst-run Ghidra12.1.4; not REA Evidence. Address "+args[i]+" */\n"+result.getDecompiledFunction().getC();
    Files.write(output.resolve(args[i]+".fallback.c"),text.getBytes(StandardCharsets.UTF_8));
    println("RECOVERED "+args[i]+" "+text.length());
   }else println("FAILED "+args[i]+" "+result.getErrorMessage());
  }
  decomp.dispose();
 }
}
