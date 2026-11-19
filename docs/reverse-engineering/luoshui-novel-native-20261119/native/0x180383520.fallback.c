/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x180383520 */

code * FUN_180383520(longlong param_1,longlong param_2)

{
  bool bVar1;
  char cVar2;
  code *pcVar3;
  code *pcVar4;
  code *pcVar5;
  
  pcVar3 = *(code **)(param_1 + 8);
  if (pcVar3 == PyUnicode_Type_exref) {
    pcVar3 = (code *)FUN_180383470();
    return pcVar3;
  }
  bVar1 = false;
  pcVar5 = PyUnicode_Type_exref;
  cVar2 = FUN_180382640(PyUnicode_Type_exref,pcVar3);
  if ((cVar2 != '\0') && (*(code **)(pcVar5 + 200) != (code *)0x0)) {
    bVar1 = true;
    pcVar4 = (code *)(**(code **)(pcVar5 + 200))(param_2,param_1,2);
    pcVar5 = PyUnicode_Type_exref;
    if (pcVar4 != _Py_NotImplementedStruct_exref) {
      return pcVar4;
    }
  }
  if ((((*(code **)(pcVar3 + 200) == (code *)0x0) ||
       (pcVar3 = (code *)(**(code **)(pcVar3 + 200))(param_1,param_2,2),
       pcVar5 = PyUnicode_Type_exref, pcVar3 == _Py_NotImplementedStruct_exref)) &&
      ((bVar1 || ((*(code **)(pcVar5 + 200) == (code *)0x0 ||
                  (pcVar3 = (code *)(**(code **)(pcVar5 + 200))(param_2,param_1,2),
                  pcVar3 == _Py_NotImplementedStruct_exref)))))) &&
     (pcVar3 = _Py_FalseStruct_exref, param_1 == param_2)) {
    pcVar3 = _Py_TrueStruct_exref;
  }
  return pcVar3;
}

