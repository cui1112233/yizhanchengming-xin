/* Analyst-run Ghidra12.1.4; not REA Evidence. Address 0x1801c29d0 */

int * FUN_1801c29d0(longlong param_1,undefined8 param_2,undefined8 *param_3)

{
  longlong *plVar1;
  char cVar2;
  int iVar3;
  longlong *plVar4;
  longlong *plVar5;
  longlong lVar6;
  longlong *plVar7;
  longlong *plVar8;
  longlong *plVar9;
  code *pcVar10;
  code *pcVar11;
  longlong *plVar12;
  int *piVar13;
  undefined8 uVar14;
  longlong *local_res18;
  undefined8 local_98;
  longlong *local_90;
  longlong *local_88;
  longlong *local_80;
  longlong *local_78;
  longlong *local_70;
  longlong *local_68;
  longlong *local_60;
  longlong *local_58;
  code *local_50;
  longlong *local_48;
  longlong *local_40;
  
  pcVar10 = (code *)*param_3;
  plVar9 = (longlong *)0x0;
  local_48 = (longlong *)0x0;
  local_40 = (longlong *)0x0;
  local_50 = pcVar10;
  plVar4 = (longlong *)FUN_18036d6d0();
  local_80 = plVar4;
  plVar5 = (longlong *)FUN_18036d6d0();
  plVar1 = DAT_1803bbdc0;
  local_res18 = (longlong *)0x0;
  local_70 = (longlong *)0x0;
  local_60 = (longlong *)0x0;
  local_78 = (longlong *)0x0;
  local_68 = (longlong *)0x0;
  local_88 = plVar5;
  if (DAT_1803bbdc0 == (longlong *)0x0) {
LAB_1801c2a67:
    DAT_1803bbdc0 = (longlong *)FUN_18036cd10(param_1,DAT_1803bc4a8,DAT_1803ca610,0x30);
  }
  else if ((1 < *DAT_1803bbdc0) || (DAT_1803bbdc0[2] != 0)) {
    if ((DAT_1803bbdc0 != (longlong *)0x0) &&
       ((-1 < (int)*DAT_1803bbdc0 && (*DAT_1803bbdc0 = *DAT_1803bbdc0 + -1, *plVar1 == 0)))) {
      (**(code **)(plVar1[1] + 0x30))();
    }
    goto LAB_1801c2a67;
  }
  plVar1 = DAT_1803bbdc0;
  FUN_180001d00(param_1,DAT_1803bbdc0);
  if (pcVar10 != _Py_TrueStruct_exref) {
    if ((pcVar10 != _Py_FalseStruct_exref) && (pcVar10 != _Py_NoneStruct_exref)) {
      lVar6 = *(longlong *)(pcVar10 + 8);
      if ((*(longlong *)(lVar6 + 0x60) == 0) ||
         (pcVar11 = *(code **)(*(longlong *)(lVar6 + 0x60) + 0x48), pcVar11 == (code *)0x0)) {
        if ((*(undefined8 **)(lVar6 + 0x70) == (undefined8 *)0x0) ||
           (pcVar11 = (code *)**(undefined8 **)(lVar6 + 0x70), pcVar11 == (code *)0x0)) {
          if ((*(undefined8 **)(lVar6 + 0x68) == (undefined8 *)0x0) ||
             (pcVar11 = (code *)**(undefined8 **)(lVar6 + 0x68), pcVar11 == (code *)0x0))
          goto LAB_1801c2b41;
          lVar6 = (*pcVar11)(pcVar10);
        }
        else {
          lVar6 = (*pcVar11)(pcVar10);
        }
      }
      else {
        iVar3 = (*pcVar11)(pcVar10);
        lVar6 = (longlong)iVar3;
      }
      if (0 < lVar6) goto LAB_1801c2b41;
      if (lVar6 != 0) {
        plVar5 = *(longlong **)(param_1 + 0x60);
        uVar14 = 0x62;
        *(undefined8 *)(param_1 + 0x60) = 0;
        local_res18 = plVar9;
        goto LAB_1801c3f5b;
      }
    }
/* novels.prompts.inkos.reviser_prompts#20="allow-full" */
    piVar13 = DAT_1803bbeb0;
    local_res18 = plVar9;
/* novels.prompts.inkos.reviser_prompts#20="allow-full" */
    if (*DAT_1803bbeb0 + 1 != 0) {
/* novels.prompts.inkos.reviser_prompts#20="allow-full" */
      *DAT_1803bbeb0 = *DAT_1803bbeb0 + 1;
    }
LAB_1801c4184:
    lVar6 = **(longlong **)(param_1 + 0x38);
    plVar1 = *(longlong **)(lVar6 + 0x30);
    **(longlong **)(param_1 + 0x38) = *(longlong *)(lVar6 + 8);
    *(undefined4 *)(plVar1 + 8) = 0xffffffff;
    plVar9 = (longlong *)plVar1[2];
    if (((plVar9 != (longlong *)0x0) && (plVar1[2] = 0, -1 < (int)*plVar9)) &&
       (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
      (**(code **)(plVar9[1] + 0x30))();
    }
    if ((-1 < (int)*plVar1) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)) {
      (**(code **)(plVar1[1] + 0x30))(plVar1);
    }
    plVar1[10] = 0;
LAB_1801c41f5:
    if (((local_48 != (longlong *)0x0) && (-1 < (int)*local_48)) &&
       (*local_48 = *local_48 + -1, *local_48 == 0)) {
      (**(code **)(local_48[1] + 0x30))();
    }
    if (((local_40 != (longlong *)0x0) && (-1 < (int)*local_40)) &&
       (*local_40 = *local_40 + -1, *local_40 == 0)) {
      (**(code **)(local_40[1] + 0x30))();
    }
    if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
      (**(code **)(plVar4[1] + 0x30))(plVar4);
    }
    if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
      (**(code **)(plVar5[1] + 0x30))(plVar5);
    }
    if (((local_res18 != (longlong *)0x0) && (-1 < (int)*local_res18)) &&
       (*local_res18 = *local_res18 + -1, *local_res18 == 0)) {
      (**(code **)(local_res18[1] + 0x30))(local_res18);
    }
    if (*(int *)local_50 < 0) {
      return piVar13;
    }
    *(longlong *)local_50 = *(longlong *)local_50 + -1;
    if (*(longlong *)local_50 != 0) {
      return piVar13;
    }
    (**(code **)(*(longlong *)(local_50 + 8) + 0x30))(local_50);
    return piVar13;
  }
LAB_1801c2b41:
  plVar4 = (longlong *)FUN_18000d280();
  if (plVar4 == (longlong *)0x0) {
    plVar5 = *(longlong **)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
    uVar14 = 100;
    local_res18 = (longlong *)0x0;
  }
  else {
    local_58 = (longlong *)FUN_180372160(param_1,0);
    do {
      pcVar10 = *(code **)(plVar4[1] + 0xe0);
      if (pcVar10 == (code *)0x0) {
        PyErr_Format(*(undefined8 *)PyExc_TypeError_exref,"\'%s\' object is not an iterator",
                     *(undefined8 *)(plVar4[1] + 0x18));
LAB_1801c2f5f:
        plVar5 = *(longlong **)(param_1 + 0x60);
        if (plVar5 != (longlong *)0x0) {
          plVar7 = (longlong *)plVar5[1];
          if ((*(uint *)((longlong *)plVar5[1] + 0x15) & 0x40000000) == 0) {
            plVar7 = plVar5;
          }
          if ((plVar7 != *(longlong **)PyExc_StopIteration_exref) &&
             (((-1 < *(int *)(plVar7[1] + 0xa8) || ((*(uint *)(plVar7 + 0x15) & 0x40000000) == 0))
              || (cVar2 = FUN_180382640(), cVar2 != '\x01')))) goto LAB_1801c2fb7;
          *(undefined8 *)(param_1 + 0x60) = 0;
          if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
            (**(code **)(plVar5[1] + 0x30))(plVar5);
          }
        }
        plVar5 = local_58;
        iVar3 = (int)*local_58 + 1;
        if (iVar3 != 0) {
          *(int *)local_58 = iVar3;
        }
        if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
          (**(code **)(plVar4[1] + 0x30))(plVar4);
        }
        if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
          (**(code **)(plVar5[1] + 0x30))(plVar5);
        }
        if (((plVar9 != (longlong *)0x0) && (-1 < (int)*plVar9)) &&
           (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
          (**(code **)(plVar9[1] + 0x30))(plVar9);
        }
        if (((plVar9 != (longlong *)0x0) && (-1 < (int)*plVar9)) &&
           (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
          (**(code **)(plVar9[1] + 0x30))(plVar9);
        }
        local_48 = plVar5;
        iVar3 = FUN_180001510(plVar5);
        if (iVar3 == -1) {
          plVar5 = *(longlong **)(param_1 + 0x60);
          uVar14 = 0x68;
          local_res18 = (longlong *)0x0;
          *(undefined8 *)(param_1 + 0x60) = 0;
          goto LAB_1801c3f5b;
        }
        if (iVar3 == 0) goto LAB_1801c383f;
        plVar5 = (longlong *)FUN_18000d280();
        if (plVar5 == (longlong *)0x0) {
          plVar5 = *(longlong **)(param_1 + 0x60);
          uVar14 = 0x69;
          local_res18 = (longlong *)0x0;
          *(undefined8 *)(param_1 + 0x60) = 0;
          goto LAB_1801c3f5b;
        }
        local_98 = FUN_18036d750(plVar5);
        plVar4 = (longlong *)
/* novels.prompts.inkos.reviser_prompts#32="<genexpr>" ; novels.prompts.inkos.reviser_prompts#33="resolve_auto_output_mode.<locals>.<genexpr>" */
                 FUN_180367890(FUN_1801c42c0,DAT_1803ca610,DAT_1803bbf10,DAT_1803bbf18,DAT_1803bbd40
                               ,&local_98,1,0x448);
        if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
          (**(code **)(plVar5[1] + 0x30))(plVar5);
        }
        plVar5 = (longlong *)FUN_1803753b0(param_1,plVar4);
        if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
          (**(code **)(plVar4[1] + 0x30))(plVar4);
        }
        if (plVar5 == (longlong *)0x0) {
          uVar14 = 0x69;
          goto LAB_1801c3f4f;
        }
        iVar3 = FUN_180001510(plVar5);
        if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
          (**(code **)(plVar5[1] + 0x30))(plVar5);
        }
/* novels.prompts.inkos.reviser_prompts#24="rewrite-only" */
        piVar13 = DAT_1803bbed0;
        plVar4 = local_80;
        plVar5 = local_88;
        if (iVar3 == 1) goto LAB_1801c4179;
        plVar4 = (longlong *)FUN_18000d280();
        uVar14 = 0x6e;
        if (plVar4 == (longlong *)0x0) {
          plVar5 = *(longlong **)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          local_res18 = (longlong *)0x0;
          goto LAB_1801c3f5b;
        }
        plVar9 = (longlong *)FUN_180372160(param_1,0);
        goto LAB_1801c3260;
      }
      plVar7 = (longlong *)(*pcVar10)(plVar4);
      if (plVar7 == (longlong *)0x0) goto LAB_1801c2f5f;
      if (((plVar9 != (longlong *)0x0) && (-1 < (int)*plVar9)) &&
         (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
        (**(code **)(plVar9[1] + 0x30))();
      }
      iVar3 = (int)*plVar7 + 1;
      if (iVar3 != 0) {
        *(int *)plVar7 = iVar3;
      }
      if (((plVar9 != (longlong *)0x0) && (-1 < (int)*plVar9)) &&
         (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
        (**(code **)(plVar9[1] + 0x30))();
      }
      plVar5 = (longlong *)FUN_18037f0c0();
      plVar9 = plVar7;
      if (plVar5 == (longlong *)0x0) {
LAB_1801c2f32:
        plVar5 = *(longlong **)(param_1 + 0x60);
LAB_1801c2fb7:
        *(undefined8 *)(param_1 + 0x60) = 0;
        plVar7 = plVar9;
        goto LAB_1801c2fbf;
      }
      *(undefined4 *)(plVar1 + 5) = 0x66;
      plVar8 = (longlong *)FUN_180378210(param_1,plVar5,DAT_1803bbe40);
      if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
        (**(code **)(plVar5[1] + 0x30))(plVar5);
      }
      if (plVar8 == (longlong *)0x0) goto LAB_1801c2f32;
      plVar5 = (longlong *)FUN_1803749a0(plVar8);
      if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
        (**(code **)(plVar8[1] + 0x30))(plVar8);
      }
      if (plVar5 == (longlong *)0x0) goto LAB_1801c2f32;
      plVar8 = (longlong *)FUN_18037f0c0();
      if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
        (**(code **)(plVar5[1] + 0x30))(plVar5);
      }
      if (plVar8 == (longlong *)0x0) goto LAB_1801c2f32;
      *(undefined4 *)(plVar1 + 5) = 0x66;
      plVar5 = (longlong *)FUN_180376210(param_1,plVar8);
      if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
        (**(code **)(plVar8[1] + 0x30))(plVar8);
      }
      if (plVar5 == (longlong *)0x0) goto LAB_1801c2f32;
/* novels.prompts.inkos.reviser_prompts#21="info" */
      iVar3 = FUN_180386310(plVar5,DAT_1803bbeb8);
      if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
        (**(code **)(plVar5[1] + 0x30))(plVar5);
      }
      if (iVar3 == -1) goto LAB_1801c2f32;
      if (iVar3 == 1) {
        plVar5 = (longlong *)FUN_18037f0c0();
        if (plVar5 != (longlong *)0x0) {
          *(undefined4 *)(plVar1 + 5) = 0x66;
          plVar8 = (longlong *)FUN_1803770d0(param_1,plVar5,DAT_1803bbec0);
          if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
            (**(code **)(plVar5[1] + 0x30))(plVar5);
          }
          if (plVar8 != (longlong *)0x0) {
            iVar3 = FUN_180001510(plVar8);
            if (iVar3 == -1) goto LAB_1801c2f01;
            if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
              (**(code **)(plVar8[1] + 0x30))(plVar8);
            }
            if (iVar3 == 0) goto LAB_1801c2dab;
LAB_1801c2e36:
            plVar5 = local_58;
            lVar6 = local_58[2];
            cVar2 = FUN_1803722b0(local_58,lVar6 + 1);
            if (cVar2 != '\0') {
              iVar3 = (int)*plVar7 + 1;
              if (iVar3 != 0) {
                *(int *)plVar7 = iVar3;
              }
              *(longlong **)(plVar5[3] + lVar6 * 8) = plVar7;
              goto LAB_1801c2e65;
            }
          }
        }
        goto LAB_1801c2f32;
      }
LAB_1801c2dab:
      plVar5 = (longlong *)FUN_18037f0c0();
      if (plVar5 == (longlong *)0x0) goto LAB_1801c2f32;
      *(undefined4 *)(plVar1 + 5) = 0x66;
      plVar8 = (longlong *)FUN_1803770d0(param_1,plVar5,DAT_1803bbec8);
      if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
        (**(code **)(plVar5[1] + 0x30))(plVar5);
      }
      if (plVar8 == (longlong *)0x0) goto LAB_1801c2f32;
      iVar3 = FUN_180001510(plVar8);
      if (iVar3 == -1) {
LAB_1801c2f01:
        plVar5 = *(longlong **)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
        if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
          (**(code **)(plVar8[1] + 0x30))(plVar8);
        }
        goto LAB_1801c2fbf;
      }
      if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
        (**(code **)(plVar8[1] + 0x30))(plVar8);
      }
      if (iVar3 != 0) goto LAB_1801c2e36;
LAB_1801c2e65:
      lVar6 = *(longlong *)(param_1 + 0x10);
      if ((((*(int *)(*(longlong *)(lVar6 + 0x60) + 0x4d4) != 0) || (*(int *)(lVar6 + 0x1a0) != 0))
          && (iVar3 = Py_MakePendingCalls(), iVar3 < 0)) && (*(longlong *)(param_1 + 0x60) != 0)) {
        plVar5 = *(longlong **)(param_1 + 0x60);
        *(undefined8 *)(param_1 + 0x60) = 0;
        goto LAB_1801c2fbf;
      }
      if (*(int *)(lVar6 + 0x174) != 0) {
        PyEval_SaveThread();
        PyEval_AcquireThread(param_1);
      }
    } while (*(longlong *)(param_1 + 0x80) == 0);
    plVar5 = *(longlong **)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x80) = 0;
    uVar14 = FUN_180373490(param_1);
    *(undefined8 *)(param_1 + 0x60) = uVar14;
    if (((plVar5 != (longlong *)0x0) && (-1 < (int)*plVar5)) &&
       (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
      (**(code **)(plVar5[1] + 0x30))(plVar5);
    }
    plVar5 = *(longlong **)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
LAB_1801c2fbf:
    if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
      (**(code **)(plVar4[1] + 0x30))(plVar4);
    }
    if ((-1 < (int)*local_58) && (*local_58 = *local_58 + -1, *local_58 == 0)) {
      (**(code **)(local_58[1] + 0x30))();
    }
    if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) &&
       (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
      (**(code **)(plVar7[1] + 0x30))(plVar7);
    }
    if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) &&
       (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
      (**(code **)(plVar7[1] + 0x30))(plVar7);
    }
    uVar14 = 100;
  }
  goto LAB_1801c3f5b;
  while( true ) {
    plVar5 = (longlong *)(*pcVar10)(plVar4);
    if (plVar5 == (longlong *)0x0) goto LAB_1801c34f3;
    local_78 = plVar5;
    if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) &&
       (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
      (**(code **)(plVar7[1] + 0x30))(plVar7);
    }
    iVar3 = (int)*plVar5 + 1;
    if (iVar3 != 0) {
      *(int *)plVar5 = iVar3;
    }
    local_70 = plVar5;
    if (((plVar8 != (longlong *)0x0) && (-1 < (int)*plVar8)) &&
       (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
      (**(code **)(plVar8[1] + 0x30))();
    }
    plVar7 = (longlong *)FUN_18037f0c0();
    if (plVar7 == (longlong *)0x0) goto LAB_1801c34cd;
    *(undefined4 *)(plVar1 + 5) = 0x6e;
    plVar8 = (longlong *)FUN_180378210(param_1,plVar7,DAT_1803bbe40);
    if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
      (**(code **)(plVar7[1] + 0x30))(plVar7);
    }
    if (plVar8 == (longlong *)0x0) goto LAB_1801c34cd;
    plVar7 = (longlong *)FUN_1803749a0(plVar8);
    if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
      (**(code **)(plVar8[1] + 0x30))(plVar8);
    }
    if (plVar7 == (longlong *)0x0) goto LAB_1801c34cd;
    plVar8 = (longlong *)FUN_18037f0c0();
    if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
      (**(code **)(plVar7[1] + 0x30))(plVar7);
    }
    if (plVar8 == (longlong *)0x0) goto LAB_1801c34cd;
    *(undefined4 *)(plVar1 + 5) = 0x6e;
    plVar7 = (longlong *)FUN_180376210(param_1,plVar8);
    if ((-1 < (int)*plVar8) && (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
      (**(code **)(plVar8[1] + 0x30))(plVar8);
    }
    if (plVar7 == (longlong *)0x0) goto LAB_1801c34cd;
/* novels.prompts.inkos.reviser_prompts#21="info" */
    iVar3 = FUN_180386310(plVar7,DAT_1803bbeb8);
    if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
      (**(code **)(plVar7[1] + 0x30))(plVar7);
    }
    if (iVar3 == -1) goto LAB_1801c34cd;
    if (iVar3 == 1) {
      lVar6 = plVar9[2];
      cVar2 = FUN_1803722b0(plVar9,lVar6 + 1);
      if (cVar2 == '\0') goto LAB_1801c34cd;
      iVar3 = (int)*plVar5 + 1;
      if (iVar3 != 0) {
        *(int *)plVar5 = iVar3;
      }
      *(longlong **)(plVar9[3] + lVar6 * 8) = plVar5;
    }
    lVar6 = *(longlong *)(param_1 + 0x10);
    if ((((*(int *)(*(longlong *)(lVar6 + 0x60) + 0x4d4) != 0) || (*(int *)(lVar6 + 0x1a0) != 0)) &&
        (iVar3 = Py_MakePendingCalls(), iVar3 < 0)) &&
       (plVar5 = *(longlong **)(param_1 + 0x60), plVar5 != (longlong *)0x0)) goto LAB_1801c354b;
    if (*(int *)(lVar6 + 0x174) != 0) {
      PyEval_SaveThread();
      PyEval_AcquireThread(param_1);
    }
    if (*(longlong *)(param_1 + 0x80) != 0) break;
LAB_1801c3260:
    plVar8 = local_70;
    plVar7 = local_78;
    pcVar10 = *(code **)(plVar4[1] + 0xe0);
    if (pcVar10 == (code *)0x0) {
      PyErr_Format(*(undefined8 *)PyExc_TypeError_exref,"\'%s\' object is not an iterator",
                   *(undefined8 *)(plVar4[1] + 0x18));
LAB_1801c34f3:
      plVar5 = *(longlong **)(param_1 + 0x60);
      if (plVar5 != (longlong *)0x0) {
        plVar12 = (longlong *)plVar5[1];
        if ((*(uint *)((longlong *)plVar5[1] + 0x15) & 0x40000000) == 0) {
          plVar12 = plVar5;
        }
        if ((plVar12 != *(longlong **)PyExc_StopIteration_exref) &&
           (((-1 < *(int *)(plVar12[1] + 0xa8) || ((*(uint *)(plVar12 + 0x15) & 0x40000000) == 0))
            || (cVar2 = FUN_180382640(), cVar2 != '\x01')))) goto LAB_1801c354b;
        *(undefined8 *)(param_1 + 0x60) = 0;
        if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
          (**(code **)(plVar5[1] + 0x30))(plVar5);
        }
      }
      iVar3 = (int)*plVar9 + 1;
      if (iVar3 != 0) {
        *(int *)plVar9 = iVar3;
      }
      if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
        (**(code **)(plVar4[1] + 0x30))(plVar4);
      }
      if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
        (**(code **)(plVar9[1] + 0x30))(plVar9);
      }
      if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) &&
         (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
        (**(code **)(plVar7[1] + 0x30))(plVar7);
      }
      if (((plVar8 != (longlong *)0x0) && (-1 < (int)*plVar8)) &&
         (*plVar8 = *plVar8 + -1, *plVar8 == 0)) {
        (**(code **)(plVar8[1] + 0x30))(plVar8);
      }
      local_40 = plVar9;
      lVar6 = FUN_180373860(local_58);
      if (((-1 < lVar6) || (*(longlong *)(param_1 + 0x60) == 0)) &&
         (plVar4 = (longlong *)PyLong_FromSsize_t(lVar6), plVar4 != (longlong *)0x0)) {
        lVar6 = FUN_180373860(plVar9);
        if (((lVar6 < 0) && (*(longlong *)(param_1 + 0x60) != 0)) ||
           (plVar5 = (longlong *)PyLong_FromSsize_t(lVar6), plVar5 == (longlong *)0x0)) {
          plVar5 = *(longlong **)(param_1 + 0x60);
          *(undefined8 *)(param_1 + 0x60) = 0;
          if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
            (**(code **)(plVar4[1] + 0x30))(plVar4);
          }
          local_res18 = (longlong *)0x0;
          uVar14 = 0x6f;
          goto LAB_1801c3f5b;
        }
        cVar2 = FUN_1803839a0(plVar4,plVar5);
        if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
          (**(code **)(plVar4[1] + 0x30))(plVar4);
        }
        if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
          (**(code **)(plVar5[1] + 0x30))(plVar5);
        }
        if (cVar2 == '\0') goto LAB_1801c383f;
        plVar5 = (longlong *)FUN_18000d280();
        if (plVar5 != (longlong *)0x0) {
          local_98 = FUN_18036d750(plVar5);
          plVar4 = (longlong *)
/* novels.prompts.inkos.reviser_prompts#32="<genexpr>" ; novels.prompts.inkos.reviser_prompts#33="resolve_auto_output_mode.<locals>.<genexpr>" */
                   FUN_180367890(FUN_1801c4990,DAT_1803ca610,DAT_1803bbf10,DAT_1803bbf18,
                                 DAT_1803bbd80,&local_98,1,0x448);
          if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
            (**(code **)(plVar5[1] + 0x30))(plVar5);
          }
          plVar5 = (longlong *)FUN_180375240(param_1,plVar4);
          if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
            (**(code **)(plVar4[1] + 0x30))(plVar4);
          }
          if (plVar5 != (longlong *)0x0) {
            iVar3 = FUN_180001510(plVar5);
            if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
              (**(code **)(plVar5[1] + 0x30))(plVar5);
            }
            if (iVar3 != 1) goto LAB_1801c383f;
            local_res18 = (longlong *)0x0;
/* novels.prompts.inkos.reviser_prompts#25="patch-only" */
            piVar13 = DAT_1803bbed8;
            plVar4 = local_80;
            plVar5 = local_88;
            goto LAB_1801c4179;
          }
        }
      }
      plVar5 = *(longlong **)(param_1 + 0x60);
      local_res18 = (longlong *)0x0;
      *(undefined8 *)(param_1 + 0x60) = 0;
      uVar14 = 0x6f;
      goto LAB_1801c3f5b;
    }
  }
  plVar7 = *(longlong **)(param_1 + 0x60);
  *(undefined8 *)(param_1 + 0x80) = 0;
  plVar5 = (longlong *)FUN_180373490(param_1);
  *(longlong **)(param_1 + 0x60) = plVar5;
  if ((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) {
    *plVar7 = *plVar7 + -1;
    if (*plVar7 == 0) {
      (**(code **)(plVar7[1] + 0x30))(plVar7);
    }
LAB_1801c34cd:
    plVar5 = *(longlong **)(param_1 + 0x60);
  }
LAB_1801c354b:
  *(undefined8 *)(param_1 + 0x60) = 0;
  if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
    (**(code **)(plVar4[1] + 0x30))(plVar4);
  }
  if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
    (**(code **)(plVar9[1] + 0x30))(plVar9);
  }
  plVar4 = local_70;
  if (((local_78 != (longlong *)0x0) && (-1 < (int)*local_78)) &&
     (*local_78 = *local_78 + -1, *local_78 == 0)) {
    (**(code **)(local_78[1] + 0x30))();
    plVar4 = local_70;
  }
  goto LAB_1801c35aa;
code_r0x0001801c3b6a:
  plVar9 = *(longlong **)(param_1 + 0x60);
  *(undefined8 *)(param_1 + 0x80) = 0;
  plVar5 = (longlong *)FUN_180373490(param_1);
  *(longlong **)(param_1 + 0x60) = plVar5;
  if ((plVar9 != (longlong *)0x0) && (-1 < (int)*plVar9)) {
    *plVar9 = *plVar9 + -1;
    if (*plVar9 == 0) {
      (**(code **)(plVar9[1] + 0x30))(plVar9);
    }
LAB_1801c3bad:
    plVar5 = *(longlong **)(param_1 + 0x60);
  }
LAB_1801c3c27:
  *(undefined8 *)(param_1 + 0x60) = 0;
  if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
    (**(code **)(plVar4[1] + 0x30))(plVar4);
  }
  if ((-1 < (int)*local_res18) && (*local_res18 = *local_res18 + -1, *local_res18 == 0)) {
    (**(code **)(local_res18[1] + 0x30))(local_res18);
  }
  plVar4 = local_60;
  if (((local_68 != (longlong *)0x0) && (-1 < (int)*local_68)) &&
     (*local_68 = *local_68 + -1, *local_68 == 0)) {
    (**(code **)(local_68[1] + 0x30))();
    plVar4 = local_60;
  }
LAB_1801c35aa:
  if (((plVar4 != (longlong *)0x0) && (-1 < (int)*plVar4)) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)
     ) {
    (**(code **)(plVar4[1] + 0x30))();
  }
  local_res18 = (longlong *)0x0;
  goto LAB_1801c3f5b;
LAB_1801c4179:
  if (*piVar13 + 1 != 0) {
    *piVar13 = *piVar13 + 1;
  }
  goto LAB_1801c4184;
LAB_1801c383f:
  uVar14 = FUN_18036fb80(param_1,DAT_1803bbee0);
/* novels.prompts.inkos.reviser_prompts#27="is_structural" ; novels.prompts.inkos.reviser_prompts#28="resolve_auto_output_mode.<locals>.is_structural" */
  lVar6 = FUN_1803609a0(FUN_1801c5060,DAT_1803bbee8,DAT_1803bbef0,DAT_1803bbde0,0,0,uVar14,
                        DAT_1803ca610,0,0,0);
  local_80[2] = lVar6;
  uVar14 = FUN_18036fb80(param_1,DAT_1803bbee0);
/* novels.prompts.inkos.reviser_prompts#29="is_local" ; novels.prompts.inkos.reviser_prompts#30="resolve_auto_output_mode.<locals>.is_local" */
  lVar6 = FUN_1803609a0(FUN_1801c5cd0,DAT_1803bbef8,DAT_1803bbf00,DAT_1803bbdf0,0,0,uVar14,
                        DAT_1803ca610,0,0,0);
  local_88[2] = lVar6;
  plVar4 = (longlong *)FUN_18000d280(local_88,local_50);
  uVar14 = 0x7c;
  if (plVar4 == (longlong *)0x0) {
    plVar5 = *(longlong **)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
    local_res18 = (longlong *)0x0;
  }
  else {
    local_res18 = (longlong *)FUN_180372160(param_1,0);
    while( true ) {
      plVar5 = local_60;
      plVar9 = local_68;
      pcVar10 = *(code **)(plVar4[1] + 0xe0);
      if (pcVar10 == (code *)0x0) break;
      plVar7 = (longlong *)(*pcVar10)(plVar4);
      if (plVar7 == (longlong *)0x0) goto LAB_1801c3bd3;
      local_68 = plVar7;
      if (((plVar9 != (longlong *)0x0) && (-1 < (int)*plVar9)) &&
         (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
        (**(code **)(plVar9[1] + 0x30))(plVar9);
      }
      iVar3 = (int)*plVar7 + 1;
      if (iVar3 != 0) {
        *(int *)plVar7 = iVar3;
      }
      local_60 = plVar7;
      if (((plVar5 != (longlong *)0x0) && (-1 < (int)*plVar5)) &&
         (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
        (**(code **)(plVar5[1] + 0x30))();
      }
      plVar5 = (longlong *)FUN_18037f0c0();
      if (plVar5 == (longlong *)0x0) goto LAB_1801c3bad;
      *(undefined4 *)(plVar1 + 5) = 0x7c;
      plVar9 = (longlong *)FUN_180378210(param_1,plVar5,DAT_1803bbe40);
      if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
        (**(code **)(plVar5[1] + 0x30))(plVar5);
      }
      if (plVar9 == (longlong *)0x0) goto LAB_1801c3bad;
      plVar5 = (longlong *)FUN_1803749a0(plVar9);
      if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
        (**(code **)(plVar9[1] + 0x30))(plVar9);
      }
      if (plVar5 == (longlong *)0x0) goto LAB_1801c3bad;
      plVar9 = (longlong *)FUN_18037f0c0();
      if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
        (**(code **)(plVar5[1] + 0x30))(plVar5);
      }
      if (plVar9 == (longlong *)0x0) goto LAB_1801c3bad;
      *(undefined4 *)(plVar1 + 5) = 0x7c;
      plVar5 = (longlong *)FUN_180376210(param_1,plVar9);
      if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
        (**(code **)(plVar9[1] + 0x30))(plVar9);
      }
      if (plVar5 == (longlong *)0x0) goto LAB_1801c3bad;
/* novels.prompts.inkos.reviser_prompts#21="info" */
      iVar3 = FUN_180386310(plVar5,DAT_1803bbeb8);
      if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
        (**(code **)(plVar5[1] + 0x30))(plVar5);
      }
      if (iVar3 == -1) goto LAB_1801c3bad;
      if (iVar3 == 1) {
        lVar6 = local_res18[2];
        cVar2 = FUN_1803722b0(local_res18,lVar6 + 1);
        if (cVar2 == '\0') goto LAB_1801c3bad;
        iVar3 = (int)*plVar7 + 1;
        if (iVar3 != 0) {
          *(int *)plVar7 = iVar3;
        }
        *(longlong **)(local_res18[3] + lVar6 * 8) = plVar7;
      }
      lVar6 = *(longlong *)(param_1 + 0x10);
      if ((((*(int *)(*(longlong *)(lVar6 + 0x60) + 0x4d4) != 0) || (*(int *)(lVar6 + 0x1a0) != 0))
          && (iVar3 = Py_MakePendingCalls(), iVar3 < 0)) &&
         (plVar5 = *(longlong **)(param_1 + 0x60), plVar5 != (longlong *)0x0)) goto LAB_1801c3c27;
      if (*(int *)(lVar6 + 0x174) != 0) {
        PyEval_SaveThread();
        PyEval_AcquireThread(param_1);
      }
      if (*(longlong *)(param_1 + 0x80) != 0) goto code_r0x0001801c3b6a;
    }
    PyErr_Format(*(undefined8 *)PyExc_TypeError_exref,"\'%s\' object is not an iterator",
                 *(undefined8 *)(plVar4[1] + 0x18));
LAB_1801c3bd3:
    plVar5 = *(longlong **)(param_1 + 0x60);
    if (plVar5 != (longlong *)0x0) {
      plVar7 = (longlong *)plVar5[1];
      if ((*(uint *)((longlong *)plVar5[1] + 0x15) & 0x40000000) == 0) {
        plVar7 = plVar5;
      }
      if ((plVar7 != *(longlong **)PyExc_StopIteration_exref) &&
         (((-1 < *(int *)(plVar7[1] + 0xa8) || ((*(uint *)(plVar7 + 0x15) & 0x40000000) == 0)) ||
          (cVar2 = FUN_180382640(), cVar2 != '\x01')))) goto LAB_1801c3c27;
      *(undefined8 *)(param_1 + 0x60) = 0;
      if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
        (**(code **)(plVar5[1] + 0x30))(plVar5);
      }
    }
    iVar3 = (int)*local_res18 + 1;
    if (iVar3 != 0) {
      *(int *)local_res18 = iVar3;
    }
    if ((-1 < (int)*plVar4) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)) {
      (**(code **)(plVar4[1] + 0x30))(plVar4);
    }
    if ((-1 < (int)*local_res18) && (*local_res18 = *local_res18 + -1, *local_res18 == 0)) {
      (**(code **)(local_res18[1] + 0x30))(local_res18);
    }
    if (((plVar9 != (longlong *)0x0) && (-1 < (int)*plVar9)) &&
       (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
      (**(code **)(plVar9[1] + 0x30))(plVar9);
    }
    if (((local_60 != (longlong *)0x0) && (-1 < (int)*local_60)) &&
       (*local_60 = *local_60 + -1, *local_60 == 0)) {
      (**(code **)(local_60[1] + 0x30))();
    }
    iVar3 = FUN_180001510(local_res18);
/* novels.prompts.inkos.reviser_prompts#25="patch-only" */
    piVar13 = DAT_1803bbed8;
    if (iVar3 == -1) {
      plVar5 = *(longlong **)(param_1 + 0x60);
      uVar14 = 0x7d;
      *(undefined8 *)(param_1 + 0x60) = 0;
      goto LAB_1801c3f5b;
    }
    if (iVar3 == 0) {
      plVar4 = local_80;
      plVar5 = local_88;
/* novels.prompts.inkos.reviser_prompts#25="patch-only" */
      if (*DAT_1803bbed8 + 1 != 0) {
/* novels.prompts.inkos.reviser_prompts#25="patch-only" */
        *DAT_1803bbed8 = *DAT_1803bbed8 + 1;
      }
      goto LAB_1801c4184;
    }
    plVar5 = (longlong *)FUN_18000d280();
    if (plVar5 == (longlong *)0x0) {
      plVar5 = *(longlong **)(param_1 + 0x60);
      uVar14 = 0x7f;
      *(undefined8 *)(param_1 + 0x60) = 0;
      goto LAB_1801c3f5b;
    }
    local_98 = FUN_18036d750(plVar5);
    plVar4 = local_80;
    local_90 = local_80;
    iVar3 = (int)*local_80 + 1;
    if (iVar3 != 0) {
      *(int *)local_80 = iVar3;
    }
    plVar9 = (longlong *)
/* novels.prompts.inkos.reviser_prompts#32="<genexpr>" ; novels.prompts.inkos.reviser_prompts#33="resolve_auto_output_mode.<locals>.<genexpr>" */
             FUN_180367890(FUN_1801c6940,DAT_1803ca610,DAT_1803bbf10,DAT_1803bbf18,DAT_1803bbd60,
                           &local_98,2,0x448);
    if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
      (**(code **)(plVar5[1] + 0x30))(plVar5);
    }
    plVar5 = (longlong *)FUN_1803753b0(param_1,plVar9);
    if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
      (**(code **)(plVar9[1] + 0x30))(plVar9);
    }
    if (plVar5 == (longlong *)0x0) {
      uVar14 = 0x7f;
      goto LAB_1801c3f4f;
    }
    iVar3 = FUN_180001510(plVar5);
    if ((-1 < (int)*plVar5) && (*plVar5 = *plVar5 + -1, *plVar5 == 0)) {
      (**(code **)(plVar5[1] + 0x30))(plVar5);
    }
/* novels.prompts.inkos.reviser_prompts#24="rewrite-only" */
    piVar13 = DAT_1803bbed0;
    plVar5 = local_88;
    if (iVar3 == 1) goto LAB_1801c4179;
    plVar9 = (longlong *)FUN_18000d280();
    if (plVar9 != (longlong *)0x0) {
      local_98 = FUN_18036d750(plVar9);
      plVar5 = local_88;
      local_90 = local_88;
      iVar3 = (int)*local_88 + 1;
      if (iVar3 != 0) {
        *(int *)local_88 = iVar3;
      }
      plVar7 = (longlong *)
/* novels.prompts.inkos.reviser_prompts#32="<genexpr>" ; novels.prompts.inkos.reviser_prompts#33="resolve_auto_output_mode.<locals>.<genexpr>" */
               FUN_180367890(FUN_1801c6e10,DAT_1803ca610,DAT_1803bbf10,DAT_1803bbf18,DAT_1803bbdb0,
                             &local_98,2,0x448);
      if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
        (**(code **)(plVar9[1] + 0x30))(plVar9);
      }
      plVar9 = (longlong *)FUN_180375240(param_1,plVar7);
      if ((-1 < (int)*plVar7) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)) {
        (**(code **)(plVar7[1] + 0x30))(plVar7);
      }
      if (plVar9 != (longlong *)0x0) {
        iVar3 = FUN_180001510(plVar9);
        if ((-1 < (int)*plVar9) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)) {
          (**(code **)(plVar9[1] + 0x30))(plVar9);
        }
/* novels.prompts.inkos.reviser_prompts#25="patch-only" */
        piVar13 = DAT_1803bbed8;
        if (iVar3 == 1) goto LAB_1801c4179;
        FUN_180001d80(param_1);
/* novels.prompts.inkos.reviser_prompts#20="allow-full" */
        piVar13 = DAT_1803bbeb0;
/* novels.prompts.inkos.reviser_prompts#20="allow-full" */
        if (*DAT_1803bbeb0 + 1 != 0) {
/* novels.prompts.inkos.reviser_prompts#20="allow-full" */
          *DAT_1803bbeb0 = *DAT_1803bbeb0 + 1;
        }
        goto LAB_1801c41f5;
      }
    }
    uVar14 = 0x81;
LAB_1801c3f4f:
    plVar5 = *(longlong **)(param_1 + 0x60);
    *(undefined8 *)(param_1 + 0x60) = 0;
  }
LAB_1801c3f5b:
  piVar13 = (int *)plVar5[5];
  if (piVar13 == (int *)0x0) {
    pcVar10 = (code *)FUN_18038e2a0(plVar1,uVar14);
    plVar4 = (longlong *)plVar5[5];
    pcVar11 = pcVar10;
    if (pcVar10 == _Py_NoneStruct_exref) {
      pcVar11 = (code *)0x0;
    }
    else {
LAB_1801c3fbc:
      pcVar10 = pcVar11;
      if ((pcVar11 != (code *)0x0) && (*(int *)pcVar11 + 1 != 0)) {
        *(int *)pcVar11 = *(int *)pcVar11 + 1;
      }
    }
  }
  else {
    if (*(longlong **)(piVar13 + 6) == plVar1) goto LAB_1801c4005;
    pcVar10 = (code *)FUN_18038e2a0(plVar1,uVar14);
    *(int **)(pcVar10 + 0x10) = piVar13;
    if (*piVar13 + 1 != 0) {
      *piVar13 = *piVar13 + 1;
    }
    plVar4 = (longlong *)plVar5[5];
    pcVar11 = pcVar10;
    if (pcVar10 != _Py_NoneStruct_exref) goto LAB_1801c3fbc;
    pcVar11 = (code *)0x0;
  }
  plVar5[5] = (longlong)pcVar11;
  if (((plVar4 != (longlong *)0x0) && (-1 < (int)*plVar4)) && (*plVar4 = *plVar4 + -1, *plVar4 == 0)
     ) {
    (**(code **)(plVar4[1] + 0x30))();
  }
  if (((pcVar10 != (code *)0x0) && (-1 < *(int *)pcVar10)) &&
     (*(longlong *)pcVar10 = *(longlong *)pcVar10 + -1, *(longlong *)pcVar10 == 0)) {
    (**(code **)(*(longlong *)(pcVar10 + 8) + 0x30))(pcVar10);
  }
LAB_1801c4005:
  plVar7 = local_40;
  plVar9 = local_48;
  pcVar10 = local_50;
  FUN_18036d0a0(plVar1,"ooocco",local_50,local_48,local_40,local_80,local_88,local_res18);
  plVar4 = DAT_1803bbdc0;
  if (plVar1 == DAT_1803bbdc0) {
    if ((-1 < (int)*DAT_1803bbdc0) && (*DAT_1803bbdc0 = *DAT_1803bbdc0 + -1, *plVar4 == 0)) {
      (**(code **)(plVar4[1] + 0x30))();
    }
    DAT_1803bbdc0 = (longlong *)0x0;
  }
  FUN_180001d80(param_1);
  if (((plVar9 != (longlong *)0x0) && (-1 < (int)*plVar9)) && (*plVar9 = *plVar9 + -1, *plVar9 == 0)
     ) {
    (**(code **)(plVar9[1] + 0x30))(plVar9);
  }
  if (((plVar7 != (longlong *)0x0) && (-1 < (int)*plVar7)) && (*plVar7 = *plVar7 + -1, *plVar7 == 0)
     ) {
    (**(code **)(plVar7[1] + 0x30))(plVar7);
  }
  if ((-1 < (int)*local_80) && (*local_80 = *local_80 + -1, *local_80 == 0)) {
    (**(code **)(local_80[1] + 0x30))();
  }
  if ((-1 < (int)*local_88) && (*local_88 = *local_88 + -1, *local_88 == 0)) {
    (**(code **)(local_88[1] + 0x30))();
  }
  if (((local_res18 != (longlong *)0x0) && (-1 < (int)*local_res18)) &&
     (*local_res18 = *local_res18 + -1, *local_res18 == 0)) {
    (**(code **)(local_res18[1] + 0x30))(local_res18);
  }
  if ((-1 < *(int *)pcVar10) &&
     (*(longlong *)pcVar10 = *(longlong *)pcVar10 + -1, *(longlong *)pcVar10 == 0)) {
    (**(code **)(*(longlong *)(pcVar10 + 8) + 0x30))(pcVar10);
  }
  plVar1 = *(longlong **)(param_1 + 0x60);
  *(longlong **)(param_1 + 0x60) = plVar5;
  if (((plVar1 != (longlong *)0x0) && (-1 < (int)*plVar1)) && (*plVar1 = *plVar1 + -1, *plVar1 == 0)
     ) {
    (**(code **)(plVar1[1] + 0x30))();
  }
  return (int *)0x0;
}
