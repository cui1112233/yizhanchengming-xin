/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x1802c5b10 */

code * FUN_1802c5b10(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  longlong *plVar2;
  int *piVar3;
  longlong *plVar4;
  int iVar5;
  longlong *plVar6;
  code *pcVar7;
  longlong lVar8;
  longlong *plVar9;
  longlong *plVar10;
  code *pcVar11;
  code *pcVar12;
  undefined8 uVar13;
  longlong *local_58;
  longlong *local_50;
  longlong *local_48;
  
  plVar4 = DAT_1803c4b70;
  pcVar7 = (code *)0x0;
  plVar1 = (longlong *)*param_3;
  plVar2 = (longlong *)param_3[1];
  if (DAT_1803c4b70 == (longlong *)0x0) {
LAB_1802c5b64:
    DAT_1803c4b70 = (longlong *)FUN_18036cd10(param_1,DAT_1803c4c08,DAT_1803ca468,0x18);
  }
  else if ((1 < *DAT_1803c4b70) || (DAT_1803c4b70[2] != 0)) {
    if ((DAT_1803c4b70 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803c4b70 && (*DAT_1803c4b70 = *DAT_1803c4b70 + -1, *plVar4 == 0)))) {
      (**(code **)(plVar4[1] + 0x30))(plVar4);
    }
    goto LAB_1802c5b64;
  }
  plVar4 = DAT_1803c4b70;
  FUN_180001d00(param_1,DAT_1803c4b70);
  plVar6 = (longlong *)FUN_18037f0c0();
  uVar13 = 0x50;
  if (plVar6 == (longlong *)0x0) {
    lVar8 = *(longlong *)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
  }
  else {
    *(undefined4 *)(plVar4 + 5) = 0x50;
    pcVar7 = (code *)FUN_180376770(param_1,plVar6,plVar2);
    if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
      (**(code **)(plVar6[1] + 0x30))(plVar6);
    }
    if (pcVar7 == (code *)0x0) {
      lVar8 = *(longlong *)(param_1 + 0x60);
      pcVar7 = (code *)0x0;
      *(undefined8 *)(param_1 + 0x60) = 0;
    }
    else {
      iVar5 = FUN_180383a30(pcVar7,DAT_1803ab260);
      if (iVar5 == -1) {
        lVar8 = *(longlong *)(param_1 + 0x60);
        uVar13 = 0x51;
        *(undefined8 *)(param_1 + 0x60) = 0;
      }
      else {
        if (iVar5 == 1) {
LAB_1802c5f80:
          pcVar11 = _Py_NoneStruct_exref;
          if (*(int *)_Py_NoneStruct_exref + 1 != 0) {
            *(int *)_Py_NoneStruct_exref = *(int *)_Py_NoneStruct_exref + 1;
          }
          FUN_180001d80(param_1);
LAB_1802c5fb5:
          if ((-1 < *(int *)pcVar7) &&
             (*(longlong *)pcVar7 = *(longlong *)pcVar7 + -1, *(longlong *)pcVar7 == 0)) {
            (**(code **)(*(longlong *)(pcVar7 + 8) + 0x30))(pcVar7);
          }
          if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
            (**(code **)(plVar1[1] + 0x30))(plVar1);
          }
          if ((int)*plVar2 < 0) {
            return pcVar11;
          }
          *plVar2 = *plVar2 + -1;
          if (*plVar2 != 0) {
            return pcVar11;
          }
          (**(code **)(plVar2[1] + 0x30))(plVar2);
          return pcVar11;
        }
        plVar6 = (longlong *)FUN_18037f0c0();
        if (plVar6 == (longlong *)0x0) {
          lVar8 = *(longlong *)(param_1 + 0x60);
          uVar13 = 0x53;
          *(undefined8 *)(param_1 + 0x60) = 0;
        }
        else {
          lVar8 = FUN_180373860(plVar2);
          plVar10 = (longlong *)(param_1 + 0x60);
          if (((lVar8 < 0) && (*plVar10 != 0)) ||
             (plVar9 = (longlong *)PyLong_FromSsize_t(lVar8), plVar9 == (longlong *)0x0)) {
            lVar8 = *plVar10;
            *plVar10 = 0;
          }
          else {
            local_48 = (longlong *)FUN_180380b80(pcVar7,plVar9);
            if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
              (**(code **)(plVar9[1] + 0x30))(plVar9);
            }
            if (local_48 != (longlong *)0x0) {
              uVar13 = 0x53;
              *(undefined4 *)(plVar4 + 5) = 0x53;
              local_58 = plVar2;
              local_50 = local_48;
              plVar10 = (longlong *)FUN_180377750(param_1,plVar6,&local_58);
              if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
                (**(code **)(plVar6[1] + 0x30))(plVar6);
              }
              if ((-1 < (int)*local_48) && (*local_48 = *local_48 + -1, *local_48 == 0)) {
                (**(code **)(local_48[1] + 0x30))(local_48);
              }
              plVar6 = (longlong *)(param_1 + 0x60);
              if (plVar10 == (longlong *)0x0) {
                lVar8 = *plVar6;
                *plVar6 = 0;
              }
              else {
                iVar5 = FUN_180386510(plVar10,DAT_1803ab260);
                if ((-1 < (int)*plVar10) && (*plVar10 = *plVar10 + -1, *plVar10 == 0)) {
                  (**(code **)(plVar10[1] + 0x30))(plVar10);
                }
                if (iVar5 != -1) {
                  if (iVar5 != 1) {
                    FUN_180001d80(param_1);
                    pcVar11 = pcVar7;
                    if (*(int *)pcVar7 + 1 != 0) {
                      *(int *)pcVar7 = *(int *)pcVar7 + 1;
                    }
                    goto LAB_1802c5fb5;
                  }
                  goto LAB_1802c5f80;
                }
                lVar8 = *plVar6;
                *plVar6 = 0;
              }
              goto LAB_1802c5de8;
            }
            lVar8 = *plVar10;
            *plVar10 = 0;
          }
          if ((-1 < (int)*plVar6) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)) {
            (**(code **)(plVar6[1] + 0x30))(plVar6);
          }
          uVar13 = 0x53;
        }
      }
    }
  }
LAB_1802c5de8:
  piVar3 = *(int **)(lVar8 + 0x28);
  if (piVar3 == (int *)0x0) {
    pcVar11 = (code *)FUN_18038e2a0(plVar4,uVar13);
    plVar6 = *(longlong **)(lVar8 + 0x28);
    if (pcVar11 == _Py_NoneStruct_exref) {
      pcVar12 = (code *)0x0;
    }
    else {
      pcVar12 = pcVar11;
      if ((pcVar11 != (code *)0x0) && (*(int *)pcVar11 + 1 != 0)) {
        *(int *)pcVar11 = *(int *)pcVar11 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar3 + 6) == plVar4) goto LAB_1802c5eac;
    pcVar11 = (code *)FUN_18038e2a0(plVar4,uVar13);
    *(int **)(pcVar11 + 0x10) = piVar3;
    if (*piVar3 + 1 != 0) {
      *piVar3 = *piVar3 + 1;
    }
    plVar6 = *(longlong **)(lVar8 + 0x28);
    if (pcVar11 == _Py_NoneStruct_exref) {
      pcVar12 = (code *)0x0;
    }
    else {
      pcVar12 = pcVar11;
      if ((pcVar11 != (code *)0x0) && (*(int *)pcVar11 + 1 != 0)) {
        *(int *)pcVar11 = *(int *)pcVar11 + 1;
      }
    }
  }
  *(code **)(lVar8 + 0x28) = pcVar12;
  if (((plVar6 != (longlong *)0x0) && (-1 < (int)*plVar6)) && (*plVar6 = *plVar6 + -1, *plVar6 == 0)
     ) {
    (**(code **)(plVar6[1] + 0x30))();
  }
  if (((pcVar11 != (code *)0x0) && (-1 < *(int *)pcVar11)) &&
     (*(longlong *)pcVar11 = *(longlong *)pcVar11 + -1, *(longlong *)pcVar11 == 0)) {
    (**(code **)(*(longlong *)(pcVar11 + 8) + 0x30))(pcVar11);
  }
LAB_1802c5eac:
  FUN_18036d0a0(plVar4,&DAT_1803948e8,plVar1,plVar2,pcVar7);
  plVar6 = DAT_1803c4b70;
  if (plVar4 == DAT_1803c4b70) {
    if ((-1 < (int)*DAT_1803c4b70) && (*DAT_1803c4b70 = *DAT_1803c4b70 + -1, *plVar6 == 0)) {
      (**(code **)(plVar6[1] + 0x30))();
    }
    DAT_1803c4b70 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if (((pcVar7 != (code *)0x0) && (-1 < *(int *)pcVar7)) &&
     (*(longlong *)pcVar7 = *(longlong *)pcVar7 + -1, *(longlong *)pcVar7 == 0)) {
    (**(code **)(*(longlong *)(pcVar7 + 8) + 0x30))(pcVar7);
  }
  if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
    (**(code **)(plVar1[1] + 0x30))(plVar1);
  }
  if ((-1 < (int)*plVar2) && (*plVar2 = *plVar2 + -1, *plVar2 == 0)) {
    (**(code **)(plVar2[1] + 0x30))(plVar2);
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong *)(param_1 + 0x60) = lVar8;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return (code *)0x0;
}

