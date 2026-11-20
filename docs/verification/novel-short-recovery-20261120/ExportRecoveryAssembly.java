import ghidra.app.util.headless.HeadlessScript;
import ghidra.program.model.listing.*;
import java.nio.file.*;
import java.nio.charset.StandardCharsets;
public class ExportRecoveryAssembly extends HeadlessScript {
 public void run() throws Exception {
  String[] args=getScriptArgs();Path out=Paths.get(args[0]);Files.createDirectories(out);
  for(int i=1;i<args.length;i++){
   Function f=getFunctionAt(toAddr(args[i]));if(f==null){println("UNRESOLVED "+args[i]);continue;}
   StringBuilder b=new StringBuilder("# Analyst-run Ghidra12.1.4 existing read-only project; not new REA Evidence.\n# Function "+f.getEntryPoint()+" body "+f.getBody()+"\n");
   InstructionIterator it=currentProgram.getListing().getInstructions(f.getBody(),true);
   while(it.hasNext()){
    Instruction ins=it.next();StringBuilder hex=new StringBuilder();for(byte v:ins.getBytes())hex.append(String.format("%02x",v&255));
    b.append("0x").append(ins.getAddress()).append(" ").append(hex).append(" ").append(ins).append("\n");
   }
   Files.write(out.resolve(args[i]+".ghidra-assembly.txt"),b.toString().getBytes(StandardCharsets.UTF_8));println("EXPORTED "+args[i]);
  }
 }
}
