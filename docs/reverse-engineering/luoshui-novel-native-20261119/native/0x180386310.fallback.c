/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x180386310 */

ulonglong FUN_180386310(longlong param_1,longlong param_2)

{
  bool bVar1;
  char cVar2;
  int iVar3;
  code *pcVar4;
  code *pcVar5;
  ulonglong uVar6;
  code *pcVar7;
  
  pcVar5 = *(code **)(param_1 + 8);
  if (pcVar5 == PyUnicode_Type_exref) {
    cVar2 = FUN_180386260();
    return (ulonglong)(cVar2 != '\0');
  }
  bVar1 = false;
  pcVar7 = PyUnicode_Type_exref;
  cVar2 = FUN_180382640(PyUnicode_Type_exref,pcVar5);
  if ((cVar2 != '\0') && (*(code **)(pcVar7 + 200) != (code *)0x0)) {
    bVar1 = true;
    pcVar4 = (code *)(**(code **)(pcVar7 + 200))(param_2,param_1,3);
    pcVar7 = PyUnicode_Type_exref;
    if (pcVar4 != _Py_NotImplementedStruct_exref) goto LAB_180386389;
  }
  if ((*(code **)(pcVar5 + 200) == (code *)0x0) ||
     (pcVar4 = (code *)(**(code **)(pcVar5 + 200))(param_1,param_2,3), pcVar7 = PyUnicode_Type_exref
     , pcVar4 == _Py_NotImplementedStruct_exref)) {
    if ((bVar1) ||
       ((*(code **)(pcVar7 + 200) == (code *)0x0 ||
        (pcVar5 = (code *)(**(code **)(pcVar7 + 200))(param_2,param_1,3),
        pcVar5 == _Py_NotImplementedStruct_exref)))) {
      return (ulonglong)(param_1 != param_2);
    }
    if (pcVar5 == (code *)0x0) {
      return 0xffffffff;
    }
    iVar3 = FUN_180001510(pcVar5);
    uVar6 = (ulonglong)(iVar3 != 0);
    if (-1 < *(int *)pcVar5) {
      *(longlong *)pcVar5 = *(longlong *)pcVar5 + -1;
      if (*(longlong *)pcVar5 == 0) {
        (**(code **)(*(longlong *)(pcVar5 + 8) + 0x30))(pcVar5);
        return uVar6;
      }
      return uVar6;
    }
    return uVar6;
  }
LAB_180386389:
  if (pcVar4 == (code *)0x0) {
    return 0xffffffff;
  }
  iVar3 = FUN_180001510(pcVar4);
  uVar6 = (ulonglong)(iVar3 != 0);
  if (-1 < *(int *)pcVar4) {
    *(longlong *)pcVar4 = *(longlong *)pcVar4 + -1;
    if (*(longlong *)pcVar4 == 0) {
      (**(code **)(*(longlong *)(pcVar4 + 8) + 0x30))(pcVar4);
      return uVar6;
    }
    return uVar6;
  }
  return uVar6;
}

